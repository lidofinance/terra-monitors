// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
	"github.com/go-openapi/validate"
)

// GetTxListResultTxsTxValueMsgValueOutputsCoins get tx list result txs tx value msg value outputs coins
//
// swagger:model getTxListResult.txs.tx.value.msg.value.outputs.coins
type GetTxListResultTxsTxValueMsgValueOutputsCoins struct {

	// amount
	// Required: true
	Amount *string `json:"amount"`

	// deonm
	// Required: true
	Deonm *string `json:"deonm"`
}

// Validate validates this get tx list result txs tx value msg value outputs coins
func (m *GetTxListResultTxsTxValueMsgValueOutputsCoins) Validate(formats strfmt.Registry) error {
	var res []error

	if err := m.validateAmount(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateDeonm(formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *GetTxListResultTxsTxValueMsgValueOutputsCoins) validateAmount(formats strfmt.Registry) error {

	if err := validate.Required("amount", "body", m.Amount); err != nil {
		return err
	}

	return nil
}

func (m *GetTxListResultTxsTxValueMsgValueOutputsCoins) validateDeonm(formats strfmt.Registry) error {

	if err := validate.Required("deonm", "body", m.Deonm); err != nil {
		return err
	}

	return nil
}

// ContextValidate validates this get tx list result txs tx value msg value outputs coins based on context it is used
func (m *GetTxListResultTxsTxValueMsgValueOutputsCoins) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
	return nil
}

// MarshalBinary interface implementation
func (m *GetTxListResultTxsTxValueMsgValueOutputsCoins) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *GetTxListResultTxsTxValueMsgValueOutputsCoins) UnmarshalBinary(b []byte) error {
	var res GetTxListResultTxsTxValueMsgValueOutputsCoins
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
