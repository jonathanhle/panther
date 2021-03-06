package api

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
	"errors"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/lambda"
	jsoniter "github.com/json-iterator/go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	analysisApiClient "github.com/panther-labs/panther/api/gateway/analysis/client"
	analysisModels "github.com/panther-labs/panther/api/gateway/analysis/models"
	alertModels "github.com/panther-labs/panther/api/lambda/alerts/models"
	deliveryModels "github.com/panther-labs/panther/api/lambda/delivery/models"
	outputModels "github.com/panther-labs/panther/api/lambda/outputs/models"
	alertTable "github.com/panther-labs/panther/internal/log_analysis/alerts_api/table"
	"github.com/panther-labs/panther/pkg/testutils"
)

func TestGetAlert(t *testing.T) {
	// Mock the ddb client and table
	mockDdbClient := &testutils.DynamoDBMock{}
	alertsTableClient = &alertTable.AlertsTable{
		AlertsTableName:                    "alertTableName",
		Client:                             mockDdbClient,
		RuleIDCreationTimeIndexName:        "ruleIDCreationTimeIndexName",
		TimePartitionCreationTimeIndexName: "timePartitionCreationTimeIndexName",
	}

	alertID := aws.String("alert-id")
	timeNow := time.Now().UTC()
	outputIds := []string{"output-id-1", "output-id-2", "output-id-3"}

	alert := &deliveryModels.Alert{
		AlertID:             alertID,
		AnalysisDescription: aws.String("A test alert"),
		AnalysisID:          "Test.Analysis.ID",
		AnalysisName:        aws.String("Test Analysis Name"),
		Runbook:             aws.String("A runbook link"),
		Title:               aws.String("Test Alert"),
		RetryCount:          0,
		Tags:                []string{"test", "alert"},
		Type:                deliveryModels.RuleType,
		OutputIds:           outputIds,
		Severity:            "INFO",
		CreatedAt:           timeNow,
		Version:             aws.String("abc"),
	}

	expectedResult := &alertTable.AlertItem{
		AlertID:             aws.StringValue(alert.AlertID),
		RuleID:              alert.AnalysisID,
		RuleVersion:         aws.StringValue(alert.Version),
		RuleDisplayName:     alert.AnalysisName,
		Title:               aws.String("Test Alert"),
		DedupString:         "dedup",
		FirstEventMatchTime: timeNow,
		CreationTime:        timeNow,
		DeliveryResponses:   []*alertModels.DeliveryResponse{{}},
		Severity:            alert.Severity,
	}

	expectedGetItemRequest := &dynamodb.GetItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"id": {S: alertID},
		},
		TableName: aws.String(alertsTableClient.AlertsTableName),
	}

	item, err := dynamodbattribute.MarshalMap(expectedResult)
	require.NoError(t, err)

	mockDdbClient.On("GetItem", expectedGetItemRequest).Return(&dynamodb.GetItemOutput{Item: item}, nil)

	result, err := alertsTableClient.GetAlert(alertID)
	require.NoError(t, err)
	require.Equal(t, expectedResult, result)

	mockDdbClient.AssertExpectations(t)
}

type mockRoundTripper struct {
	http.RoundTripper
	mock.Mock
}

func (m *mockRoundTripper) RoundTrip(request *http.Request) (*http.Response, error) {
	args := m.Called(request)
	return args.Get(0).(*http.Response), args.Error(1)
}

func TestPopulateAlert(t *testing.T) {
	mockRoundTripper := &mockRoundTripper{}
	httpClient = &http.Client{Transport: mockRoundTripper}
	analysisConfig := analysisApiClient.DefaultTransportConfig().
		WithHost("host").
		WithBasePath("path")
	analysisClient = analysisApiClient.NewHTTPClientWithConfig(nil, analysisConfig)

	alertID := aws.String("alert-id")
	timeNow := time.Now().UTC()
	versionID := "version"
	analysisDisplayName := aws.String("Test Analysis Name")
	description := "A test aler"
	analysisID := "Test.Analysis.ID"
	runbook := "A runbook link"
	severity := "INFO"
	tags := []string{"test", "alert"}

	alert := &deliveryModels.Alert{
		AlertID:             alertID,
		AnalysisDescription: aws.String(description),
		AnalysisID:          analysisID,
		AnalysisName:        analysisDisplayName,
		Runbook:             aws.String(runbook),
		Title:               aws.String("Test Alert"),
		RetryCount:          0,
		Tags:                tags,
		Type:                deliveryModels.RuleType,
		OutputIds:           []string{},
		Severity:            severity,
		CreatedAt:           timeNow,
		Version:             aws.String(versionID),
		IsResent:            true,
	}

	alertItem := &alertTable.AlertItem{
		AlertID:             *alertID,
		RuleID:              analysisID,
		RuleVersion:         versionID,
		RuleDisplayName:     analysisDisplayName,
		Title:               aws.String("Test Alert"),
		DedupString:         "dedup",
		FirstEventMatchTime: timeNow,
		CreationTime:        timeNow,
		DeliveryResponses:   []*alertModels.DeliveryResponse{{}},
		Severity:            alert.Severity,
	}

	rule := &analysisModels.Rule{
		Body:               "",
		CreatedAt:          analysisModels.ModifyTime(timeNow),
		CreatedBy:          "user-id",
		DedupPeriodMinutes: 15,
		Description:        analysisModels.Description(description),
		DisplayName:        analysisModels.DisplayName(*analysisDisplayName),
		Enabled:            true,
		ID:                 analysisModels.ID(analysisID),
		LastModified:       analysisModels.ModifyTime(timeNow),
		LastModifiedBy:     "user-id",
		LogTypes:           []string{"log-type"},
		OutputIds:          []string{},
		Runbook:            analysisModels.Runbook(runbook),
		Severity:           analysisModels.Severity(severity),
		Tags:               tags,
		VersionID:          "version",
	}

	mockRoundTripper.On("RoundTrip", mock.Anything).Return(generateResponse(rule, http.StatusOK), nil).Once()
	expectedAlert, err := populateAlertData(alertItem)
	require.NoError(t, err)
	require.Equal(t, alert, expectedAlert)
	mockRoundTripper.AssertExpectations(t)
}

func generateResponse(body interface{}, httpCode int) *http.Response {
	serializedBody, _ := jsoniter.MarshalToString(body)
	return &http.Response{StatusCode: httpCode, Body: ioutil.NopCloser(strings.NewReader(serializedBody))}
}

func TestGetAlertOutputMapping(t *testing.T) {
	mockClient := &testutils.LambdaMock{}
	lambdaClient = mockClient

	alertID := aws.String("alert-id")
	outputIds := []string{"output-id-1", "output-id-2", "output-id-3"}

	alert := &deliveryModels.Alert{
		AlertID:             alertID,
		AnalysisDescription: aws.String("A test alert"),
		AnalysisID:          "Test.Analysis.ID",
		AnalysisName:        aws.String("Test Analysis Name"),
		Runbook:             aws.String("A runbook link"),
		Title:               aws.String("Test Alert"),
		RetryCount:          0,
		Tags:                []string{"test", "alert"},
		Type:                deliveryModels.RuleType,
		OutputIds:           []string{},
		Severity:            "INFO",
		CreatedAt:           time.Now().UTC(),
		Version:             aws.String("abc"),
	}

	input := &deliveryModels.DeliverAlertInput{
		AlertID:   aws.StringValue(alertID),
		OutputIds: outputIds,
	}

	outputs := []*outputModels.AlertOutput{
		{
			OutputID:           aws.String(outputIds[0]),
			OutputType:         aws.String("slack"),
			DefaultForSeverity: []*string{aws.String("INFO")},
		},
		{
			OutputID:           aws.String(outputIds[1]),
			OutputType:         aws.String("customwebhook"),
			DefaultForSeverity: []*string{aws.String("INFO"), aws.String("MEDIUM")},
		},
		{
			OutputID:           aws.String(outputIds[2]),
			OutputType:         aws.String("asana"),
			DefaultForSeverity: []*string{aws.String("INFO"), aws.String("MEDIUM"), aws.String("CRITICAL")},
		},
	}

	payload, err := jsoniter.Marshal(outputs)
	require.NoError(t, err)
	mockLambdaResponse := &lambda.InvokeOutput{Payload: payload}
	mockClient.On("Invoke", mock.Anything).Return(mockLambdaResponse, nil).Once()

	// AlertOutputMap map[*deliveryModels.Alert][]*outputModels.AlertOutput
	expectedResult := AlertOutputMap{
		alert: outputs,
	}

	// Need to expire the cache because other tests mutate this global when run in parallel
	outputsCache = &alertOutputsCache{
		RefreshInterval: time.Second * time.Duration(30),
		Expiry:          time.Now().Add(time.Minute * time.Duration(-5)),
	}
	result, err := getAlertOutputMapping(alert, input.OutputIds)
	require.NoError(t, err)

	assert.Equal(t, expectedResult, result)
}

func TestGetAlertOutputMappingError(t *testing.T) {
	mockClient := &testutils.LambdaMock{}
	lambdaClient = mockClient

	alertID := aws.String("alert-id")
	outputIds := []string{"output-id-1", "output-id-2", "output-id-3"}

	alert := &deliveryModels.Alert{
		AlertID:             alertID,
		AnalysisDescription: aws.String("A test alert"),
		AnalysisID:          "Test.Analysis.ID",
		AnalysisName:        aws.String("Test Analysis Name"),
		Runbook:             aws.String("A runbook link"),
		Title:               aws.String("Test Alert"),
		RetryCount:          0,
		Tags:                []string{"test", "alert"},
		Type:                deliveryModels.RuleType,
		OutputIds:           []string{},
		Severity:            "INFO",
		CreatedAt:           time.Now().UTC(),
		Version:             aws.String("abc"),
	}

	input := &deliveryModels.DeliverAlertInput{
		AlertID:   aws.StringValue(alertID),
		OutputIds: outputIds,
	}

	mockClient.On("Invoke", mock.Anything).Return((*lambda.InvokeOutput)(nil), errors.New("error")).Once()

	// AlertOutputMap map[*deliveryModels.Alert][]*outputModels.AlertOutput
	expectedResult := AlertOutputMap{}

	// Need to expire the cache because other tests mutate this global when run in parallel
	outputsCache = &alertOutputsCache{
		RefreshInterval: time.Second * time.Duration(30),
		Expiry:          time.Now().Add(time.Minute * time.Duration(-5)),
	}
	result, err := getAlertOutputMapping(alert, input.OutputIds)
	require.Error(t, err)

	assert.Equal(t, expectedResult, result)
}

func TestGetAlertOutputMappingInvalidOutputIds(t *testing.T) {
	mockClient := &testutils.LambdaMock{}
	lambdaClient = mockClient

	alertID := aws.String("alert-id")
	outputIds := []string{"output-id-1", "output-id-2", "output-id-3"}

	alert := &deliveryModels.Alert{
		AlertID:             alertID,
		AnalysisDescription: aws.String("A test alert"),
		AnalysisID:          "Test.Analysis.ID",
		AnalysisName:        aws.String("Test Analysis Name"),
		Runbook:             aws.String("A runbook link"),
		Title:               aws.String("Test Alert"),
		RetryCount:          0,
		Tags:                []string{"test", "alert"},
		Type:                deliveryModels.RuleType,
		OutputIds:           []string{},
		Severity:            "INFO",
		CreatedAt:           time.Now().UTC(),
		Version:             aws.String("abc"),
	}

	input := &deliveryModels.DeliverAlertInput{
		AlertID:   aws.StringValue(alertID),
		OutputIds: outputIds,
	}

	outputs := []*outputModels.AlertOutput{
		{
			OutputID:           aws.String("output-id-a"),
			OutputType:         aws.String("slack"),
			DefaultForSeverity: []*string{aws.String("INFO")},
		},
		{
			OutputID:           aws.String("output-id-b"),
			OutputType:         aws.String("customwebhook"),
			DefaultForSeverity: []*string{aws.String("INFO"), aws.String("MEDIUM")},
		},
		{
			OutputID:           aws.String("output-id-c"),
			OutputType:         aws.String("asana"),
			DefaultForSeverity: []*string{aws.String("INFO"), aws.String("MEDIUM"), aws.String("CRITICAL")},
		},
	}

	payload, err := jsoniter.Marshal(outputs)
	require.NoError(t, err)
	mockLambdaResponse := &lambda.InvokeOutput{Payload: payload}
	mockClient.On("Invoke", mock.Anything).Return(mockLambdaResponse, nil).Once()

	// AlertOutputMap map[*deliveryModels.Alert][]*outputModels.AlertOutput
	expectedResult := AlertOutputMap{}

	// Need to expire the cache because other tests mutate this global when run in parallel
	outputsCache = &alertOutputsCache{
		RefreshInterval: time.Second * time.Duration(30),
		Expiry:          time.Now().Add(time.Minute * time.Duration(-5)),
	}
	result, err := getAlertOutputMapping(alert, input.OutputIds)
	require.Error(t, err)

	assert.Equal(t, expectedResult, result)
}

func TestIntersection(t *testing.T) {
	outputIds := []string{"output-id-1", "output-id-2", "output-id-3"}
	outputs := []*outputModels.AlertOutput{
		{
			OutputID:           aws.String("output-id-1"),
			OutputType:         aws.String("slack"),
			DefaultForSeverity: []*string{aws.String("INFO")},
		},
		{
			OutputID:           aws.String("output-id-b"),
			OutputType:         aws.String("customwebhook"),
			DefaultForSeverity: []*string{aws.String("INFO"), aws.String("MEDIUM")},
		},
		{
			OutputID:           aws.String("output-id-3"),
			OutputType:         aws.String("asana"),
			DefaultForSeverity: []*string{aws.String("INFO"), aws.String("MEDIUM"), aws.String("CRITICAL")},
		},
	}

	expectedResult := []*outputModels.AlertOutput{
		{
			OutputID:           aws.String("output-id-1"),
			OutputType:         aws.String("slack"),
			DefaultForSeverity: []*string{aws.String("INFO")},
		},
		{
			OutputID:           aws.String("output-id-3"),
			OutputType:         aws.String("asana"),
			DefaultForSeverity: []*string{aws.String("INFO"), aws.String("MEDIUM"), aws.String("CRITICAL")},
		},
	}

	result := intersection(outputIds, outputs)
	assert.Equal(t, expectedResult, result)
}

func TestReturnIfFailedSuccess(t *testing.T) {
	alertID := aws.String("alert-id")
	outputIds := []string{"output-id-1", "output-id-2", "output-id-3"}
	timeNow := time.Now().UTC()
	alert := deliveryModels.Alert{
		AlertID:             alertID,
		AnalysisDescription: aws.String("A test alert"),
		AnalysisID:          "Test.Analysis.ID",
		AnalysisName:        aws.String("Test Analysis Name"),
		Runbook:             aws.String("A runbook link"),
		Title:               aws.String("Test Alert"),
		RetryCount:          0,
		Tags:                []string{"test", "alert"},
		Type:                deliveryModels.RuleType,
		OutputIds:           []string{},
		Severity:            "INFO",
		CreatedAt:           timeNow,
		Version:             aws.String("abc"),
	}
	dispatchStatuses := []DispatchStatus{
		{
			Alert:        alert,
			OutputID:     outputIds[0],
			Message:      "message",
			StatusCode:   200,
			Success:      true,
			NeedsRetry:   false,
			DispatchedAt: timeNow,
		},
		{
			Alert:        alert,
			OutputID:     outputIds[1],
			Message:      "message",
			StatusCode:   299,
			Success:      true,
			NeedsRetry:   false,
			DispatchedAt: timeNow,
		},
	}

	err := returnIfFailed(dispatchStatuses)
	require.NoError(t, err)
}

func TestReturnIfFailed(t *testing.T) {
	alertID := aws.String("alert-id")
	outputIds := []string{"output-id-1", "output-id-2", "output-id-3"}
	timeNow := time.Now().UTC()
	alert := deliveryModels.Alert{
		AlertID:             alertID,
		AnalysisDescription: aws.String("A test alert"),
		AnalysisID:          "Test.Analysis.ID",
		AnalysisName:        aws.String("Test Analysis Name"),
		Runbook:             aws.String("A runbook link"),
		Title:               aws.String("Test Alert"),
		RetryCount:          0,
		Tags:                []string{"test", "alert"},
		Type:                deliveryModels.RuleType,
		OutputIds:           []string{},
		Severity:            "INFO",
		CreatedAt:           timeNow,
		Version:             aws.String("abc"),
	}
	dispatchStatuses := []DispatchStatus{
		{
			Alert:        alert,
			OutputID:     outputIds[0],
			Message:      "message",
			StatusCode:   401,
			Success:      false,
			NeedsRetry:   false,
			DispatchedAt: timeNow,
		},
		{
			Alert:        alert,
			OutputID:     outputIds[1],
			Message:      "message",
			StatusCode:   299,
			Success:      true,
			NeedsRetry:   false,
			DispatchedAt: timeNow,
		},
	}

	err := returnIfFailed(dispatchStatuses)
	require.Error(t, err)
}
