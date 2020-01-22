// Code generated by go-swagger; DO NOT EDIT.

package operations

/**
 * Panther is a scalable, powerful, cloud-native SIEM written in Golang/React.
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

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"
	"net/http"
	"time"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	cr "github.com/go-openapi/runtime/client"
	strfmt "github.com/go-openapi/strfmt"

	models "github.com/panther-labs/panther/api/gateway/analysis/models"
)

// NewBulkUploadParams creates a new BulkUploadParams object
// with the default values initialized.
func NewBulkUploadParams() *BulkUploadParams {
	var ()
	return &BulkUploadParams{

		timeout: cr.DefaultTimeout,
	}
}

// NewBulkUploadParamsWithTimeout creates a new BulkUploadParams object
// with the default values initialized, and the ability to set a timeout on a request
func NewBulkUploadParamsWithTimeout(timeout time.Duration) *BulkUploadParams {
	var ()
	return &BulkUploadParams{

		timeout: timeout,
	}
}

// NewBulkUploadParamsWithContext creates a new BulkUploadParams object
// with the default values initialized, and the ability to set a context for a request
func NewBulkUploadParamsWithContext(ctx context.Context) *BulkUploadParams {
	var ()
	return &BulkUploadParams{

		Context: ctx,
	}
}

// NewBulkUploadParamsWithHTTPClient creates a new BulkUploadParams object
// with the default values initialized, and the ability to set a custom HTTPClient for a request
func NewBulkUploadParamsWithHTTPClient(client *http.Client) *BulkUploadParams {
	var ()
	return &BulkUploadParams{
		HTTPClient: client,
	}
}

/*BulkUploadParams contains all the parameters to send to the API endpoint
for the bulk upload operation typically these are written to a http.Request
*/
type BulkUploadParams struct {

	/*Body*/
	Body *models.BulkUpload

	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithTimeout adds the timeout to the bulk upload params
func (o *BulkUploadParams) WithTimeout(timeout time.Duration) *BulkUploadParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the bulk upload params
func (o *BulkUploadParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the bulk upload params
func (o *BulkUploadParams) WithContext(ctx context.Context) *BulkUploadParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the bulk upload params
func (o *BulkUploadParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the bulk upload params
func (o *BulkUploadParams) WithHTTPClient(client *http.Client) *BulkUploadParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the bulk upload params
func (o *BulkUploadParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WithBody adds the body to the bulk upload params
func (o *BulkUploadParams) WithBody(body *models.BulkUpload) *BulkUploadParams {
	o.SetBody(body)
	return o
}

// SetBody adds the body to the bulk upload params
func (o *BulkUploadParams) SetBody(body *models.BulkUpload) {
	o.Body = body
}

// WriteToRequest writes these params to a swagger request
func (o *BulkUploadParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

	if err := r.SetTimeout(o.timeout); err != nil {
		return err
	}
	var res []error

	if o.Body != nil {
		if err := r.SetBodyParam(o.Body); err != nil {
			return err
		}
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}