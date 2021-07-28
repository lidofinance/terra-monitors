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

// GetTxListResultTxsTxValueMsgValueInputs get tx list result txs tx value msg value inputs
//
// swagger:model getTxListResult.txs.tx.value.msg.value.inputs
type GetTxListResultTxsTxValueMsgValueInputs struct {

	// address
	// Required: true
	Address *string `json:"address"`

	// coins
	// Required: true
	Coins []*GetTxListResultTxsTxValueMsgValueInputsCoins `json:"coins"`
}

// Validate validates this get tx list result txs tx value msg value inputs
func (m *GetTxListResultTxsTxValueMsgValueInputs) Validate(formats strfmt.Registry) error {
	var res []error

	if err := m.validateAddress(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateCoins(formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *GetTxListResultTxsTxValueMsgValueInputs) validateAddress(formats strfmt.Registry) error {

	if err := validate.Required("address", "body", m.Address); err != nil {
		return err
	}

	return nil
}

func (m *GetTxListResultTxsTxValueMsgValueInputs) validateCoins(formats strfmt.Registry) error {

	if err := validate.Required("coins", "body", m.Coins); err != nil {
		return err
	}

	for i := 0; i < len(m.Coins); i++ {
		if swag.IsZero(m.Coins[i]) { // not required
			continue
		}

		if m.Coins[i] != nil {
			if err := m.Coins[i].Validate(formats); err != nil {
				if ve, ok := err.(*errors.Validation); ok {
					return ve.ValidateName("coins" + "." + strconv.Itoa(i))
				}
				return err
			}
		}

	}

	return nil
}

// ContextValidate validate this get tx list result txs tx value msg value inputs based on the context it is used
func (m *GetTxListResultTxsTxValueMsgValueInputs) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
	var res []error

	if err := m.contextValidateCoins(ctx, formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *GetTxListResultTxsTxValueMsgValueInputs) contextValidateCoins(ctx context.Context, formats strfmt.Registry) error {

	for i := 0; i < len(m.Coins); i++ {

		if m.Coins[i] != nil {
			if err := m.Coins[i].ContextValidate(ctx, formats); err != nil {
				if ve, ok := err.(*errors.Validation); ok {
					return ve.ValidateName("coins" + "." + strconv.Itoa(i))
				}
				return err
			}
		}

	}

	return nil
}

// MarshalBinary interface implementation
func (m *GetTxListResultTxsTxValueMsgValueInputs) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *GetTxListResultTxsTxValueMsgValueInputs) UnmarshalBinary(b []byte) error {
	var res GetTxListResultTxsTxValueMsgValueInputs
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}