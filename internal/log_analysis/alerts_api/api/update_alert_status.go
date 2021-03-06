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
	"github.com/panther-labs/panther/api/lambda/alerts/models"
	"github.com/panther-labs/panther/internal/log_analysis/alerts_api/utils"
	"github.com/panther-labs/panther/pkg/gatewayapi"
)

// UpdateAlertStatus modifies an alert's attributes.
func (API) UpdateAlertStatus(input *models.UpdateAlertStatusInput) (result *models.UpdateAlertStatusOutput, err error) {
	// Run the update alert query
	alertItem, err := alertsDB.UpdateAlertStatus(input)
	if err != nil {
		return nil, err
	}

	// If there was no item from the DB, we return an empty response
	if alertItem == nil {
		return &models.UpdateAlertStatusOutput{}, nil
	}

	// Marshal to an alert summary
	result = utils.AlertItemToSummary(alertItem)

	gatewayapi.ReplaceMapSliceNils(result)
	return result, nil
}
