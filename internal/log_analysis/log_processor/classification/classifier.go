package classification

/**
 * Panther is a Cloud-Native SIEM for the Modern Security Team.
 * Copyright (C) 2020 Panther Labs Inc
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as
 * published by the Free Software Foundation, either version 3 of the
 * License, or (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

import (
	"container/heap"
	"strings"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/panther-labs/panther/internal/log_analysis/log_processor/parsers"
)

// ClassifierAPI is the interface for a classifier
type ClassifierAPI interface {
	// Classify attempts to classify the provided log line
	Classify(log string) *ClassifierResult
	// aggregate stats
	Stats() *ClassifierStats
	// per-parser stats, map of LogType -> stats
	ParserStats() map[string]*ParserStats
}

// ClassifierResult is the result of the ClassifierAPI#Classify method
type ClassifierResult struct {
	// Events contains the parsed events
	// If the classification process was not successful and the log is from an
	// unsupported type, this will be nil
	Events []*parsers.Result
	// LogType is the identified type of the log
	LogType *string
	Stream  parsers.ResultStream
}

// NewClassifier returns a new instance of a ClassifierAPI implementation
func NewClassifier(parsers map[string]parsers.Interface) ClassifierAPI {
	return &Classifier{
		parsers:     NewParserPriorityQueue(parsers),
		parserStats: make(map[string]*ParserStats),
	}
}

// Classifier is the struct responsible for classifying logs
type Classifier struct {
	parsers *ParserPriorityQueue
	// aggregate stats
	stats ClassifierStats
	// per-parser stats, map of LogType -> stats
	parserStats map[string]*ParserStats
}

func (c *Classifier) Stats() *ClassifierStats {
	return &c.stats
}

func (c *Classifier) ParserStats() map[string]*ParserStats {
	return c.parserStats
}

// catch panics from parsers, log and continue
func safeLogParse(logType string, parser parsers.Interface, log string) (results []*parsers.Result, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errors.Errorf("parser %q panic: %v", logType, r)
			results = nil
		}
	}()
	results, err = parser.ParseLog(log)
	if err != nil {
		return nil, err
	}
	return results, nil
}

// Classify attempts to classify the provided log line
func (c *Classifier) Classify(log string) *ClassifierResult {
	startClassify := time.Now().UTC()
	// Slice containing the popped queue items
	var popped []interface{}
	result := &ClassifierResult{}

	if len(log) == 0 { // likely empty file, nothing to do
		return result
	}

	// update aggregate stats
	defer func() {
		c.stats.ClassifyTimeMicroseconds = uint64(time.Since(startClassify).Microseconds())
		c.stats.BytesProcessedCount += uint64(len(log))
		c.stats.LogLineCount++
		c.stats.EventCount += uint64(len(result.Events))
		if len(log) > 0 {
			if result.LogType == nil {
				c.stats.ClassificationFailureCount++
			} else {
				c.stats.SuccessfullyClassifiedCount++
			}
		}
	}()

	log = strings.TrimSpace(log) // often the last line has \n only, could happen mid file tho

	if len(log) == 0 { // we count above (because it is a line in the file) then skip
		return result
	}

	for c.parsers.Len() > 0 {
		currentItem := c.parsers.Peek()

		startParseTime := time.Now().UTC()
		logType := currentItem.logType
		parsedEvents, err := safeLogParse(logType, currentItem.parser, log)
		endParseTime := time.Now().UTC()

		if err == nil {
			result.LogType = &logType
			result.Events = parsedEvents

			// Since the parsing was successful, remove all penalty from the parser
			// The parser will be higher priority in the queue
			currentItem.penalty = 0
			c.updateStats(ParserStats{
				LogType:                logType,
				ParserTimeMicroseconds: uint64(endParseTime.Sub(startParseTime).Microseconds()),
				BytesProcessedCount:    uint64(len(log)),
				LogLineCount:           1,
				EventCount:             uint64(len(result.Events)),
			})
			break
		}
		// Parser error signifies streaming results
		if e, ok := err.(parsers.StreamResultsError); ok {
			stream := e.Stream()
			result.LogType = &logType
			result.Stream = &classifierStream{
				stream:     stream,
				classifier: c,
			}
			// Since the parsing was successful, remove all penalty from the parser
			// The parser will be higher priority in the queue
			currentItem.penalty = 0
			c.updateStats(ParserStats{
				LogType:                logType,
				ParserTimeMicroseconds: uint64(endParseTime.Sub(startParseTime).Microseconds()),
				BytesProcessedCount:    uint64(len(log)),
				LogLineCount:           1,
			})
			break
		}
		// Parser failed to parse event
		zap.L().Debug("failed to parse event", zap.String("expectedLogType", logType), zap.Error(err))
		// Removing parser from queue
		// Since peek takes the item with the lowest penalty, Pop should pop the correct entry.
		// It is unclear how this will behave when there are more than one parsers with the same penalty
		popped = append(popped, heap.Pop(c.parsers))
		// Increasing penalty of the parser
		// Due to increased penalty the parser will be lower priority in the queue
		currentItem.penalty++
		// record failure
	}

	// Put back the popped items to the ParserPriorityQueue.
	for _, item := range popped {
		heap.Push(c.parsers, item)
	}
	return result
}

func (c *Classifier) updateStats(stats ParserStats) {
	logType := stats.LogType
	if logType == "" {
		return
	}
	entry, ok := c.parserStats[logType]
	if !ok {
		c.parserStats[logType] = &stats
		return
	}
	entry.ParserTimeMicroseconds += stats.ParserTimeMicroseconds
	entry.BytesProcessedCount += stats.BytesProcessedCount
	entry.LogLineCount += stats.LogLineCount
	entry.EventCount += stats.EventCount
}

// aggregate stats
type ClassifierStats struct {
	ClassifyTimeMicroseconds    uint64 // total time parsing
	BytesProcessedCount         uint64 // input bytes
	LogLineCount                uint64 // input records
	EventCount                  uint64 // output records
	SuccessfullyClassifiedCount uint64
	ClassificationFailureCount  uint64
}

// per parser stats
type ParserStats struct {
	ParserTimeMicroseconds uint64 // total time parsing
	BytesProcessedCount    uint64 // input bytes
	LogLineCount           uint64 // input records
	EventCount             uint64 // output records
	LogType                string
}

type classifierStream struct {
	stream     parsers.ResultStream
	classifier *Classifier
}

func (s *classifierStream) Next() (*parsers.Result, error) {
	startTime := time.Now()
	r, err := s.stream.Next()
	if err != nil {
		return nil, err
	}
	s.classifier.updateStats(ParserStats{
		LogType:                r.LogType,
		ParserTimeMicroseconds: uint64(time.Since(startTime).Microseconds()),
		EventCount:             1,
	})
	return r, nil
}
