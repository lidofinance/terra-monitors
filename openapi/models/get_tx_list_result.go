// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"
	"strconv"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
	"github.com/go-openapi/validate"
)

// GetTxListResult get tx list result
//
// swagger:model getTxListResult
type GetTxListResult struct {

	// Per page item limit
	// Required: true
	Limit *int64 `json:"limit"`

	// next
	Next int64 `json:"next,omitempty"`

	// tx list
	// Required: true
	Txs []*GetTxListResultTxs `json:"txs"`
}

// Validate validates this get tx list result
func (m *GetTxListResult) Validate(formats strfmt.Registry) error {
	var res []error

	if err := m.validateLimit(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateTxs(formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *GetTxListResult) validateLimit(formats strfmt.Registry) error {

	if err := validate.Required("limit", "body", m.Limit); err != nil {
		return err
	}

	return nil
}

func (m *GetTxListResult) validateTxs(formats strfmt.Registry) error {

	if err := validate.Required("txs", "body", m.Txs); err != nil {
		return err
	}

	for i := 0; i < len(m.Txs); i++ {
		if swag.IsZero(m.Txs[i]) { // not required
			continue
		}

		if m.Txs[i] != nil {
			if err := m.Txs[i].Validate(formats); err != nil {
				if ve, ok := err.(*errors.Validation); ok {
					return ve.ValidateName("txs" + "." + strconv.Itoa(i))
				}
				return err
			}
		}

	}

	return nil
}

// ContextValidate validate this get tx list result based on the context it is used
func (m *GetTxListResult) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
	var res []error

	if err := m.contextValidateTxs(ctx, formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *GetTxListResult) contextValidateTxs(ctx context.Context, formats strfmt.Registry) error {

	for i := 0; i < len(m.Txs); i++ {

		if m.Txs[i] != nil {
			if err := m.Txs[i].ContextValidate(ctx, formats); err != nil {
				if ve, ok := err.(*errors.Validation); ok {
					return ve.ValidateName("txs" + "." + strconv.Itoa(i))
				}
				return err
			}
		}

	}

	return nil
}

// MarshalBinary interface implementation
func (m *GetTxListResult) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *GetTxListResult) UnmarshalBinary(b []byte) error {
	var res GetTxListResult
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
