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
)

// Block block
//
// swagger:model Block
type Block struct {

	// last commit
	LastCommit *BlockLastCommit `json:"last_commit,omitempty"`
}

// Validate validates this block
func (m *Block) Validate(formats strfmt.Registry) error {
	var res []error

	if err := m.validateLastCommit(formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *Block) validateLastCommit(formats strfmt.Registry) error {
	if swag.IsZero(m.LastCommit) { // not required
		return nil
	}

	if m.LastCommit != nil {
		if err := m.LastCommit.Validate(formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("last_commit")
			}
			return err
		}
	}

	return nil
}

// ContextValidate validate this block based on the context it is used
func (m *Block) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
	var res []error

	if err := m.contextValidateLastCommit(ctx, formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *Block) contextValidateLastCommit(ctx context.Context, formats strfmt.Registry) error {

	if m.LastCommit != nil {
		if err := m.LastCommit.ContextValidate(ctx, formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("last_commit")
			}
			return err
		}
	}

	return nil
}

// MarshalBinary interface implementation
func (m *Block) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *Block) UnmarshalBinary(b []byte) error {
	var res Block
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}

// BlockLastCommit block last commit
//
// swagger:model BlockLastCommit
type BlockLastCommit struct {

	// height
	Height string `json:"height,omitempty"`

	// signatures
	Signatures []*BlockLastCommitSignaturesItems0 `json:"signatures"`
}

// Validate validates this block last commit
func (m *BlockLastCommit) Validate(formats strfmt.Registry) error {
	var res []error

	if err := m.validateSignatures(formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *BlockLastCommit) validateSignatures(formats strfmt.Registry) error {
	if swag.IsZero(m.Signatures) { // not required
		return nil
	}

	for i := 0; i < len(m.Signatures); i++ {
		if swag.IsZero(m.Signatures[i]) { // not required
			continue
		}

		if m.Signatures[i] != nil {
			if err := m.Signatures[i].Validate(formats); err != nil {
				if ve, ok := err.(*errors.Validation); ok {
					return ve.ValidateName("last_commit" + "." + "signatures" + "." + strconv.Itoa(i))
				}
				return err
			}
		}

	}

	return nil
}

// ContextValidate validate this block last commit based on the context it is used
func (m *BlockLastCommit) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
	var res []error

	if err := m.contextValidateSignatures(ctx, formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *BlockLastCommit) contextValidateSignatures(ctx context.Context, formats strfmt.Registry) error {

	for i := 0; i < len(m.Signatures); i++ {

		if m.Signatures[i] != nil {
			if err := m.Signatures[i].ContextValidate(ctx, formats); err != nil {
				if ve, ok := err.(*errors.Validation); ok {
					return ve.ValidateName("last_commit" + "." + "signatures" + "." + strconv.Itoa(i))
				}
				return err
			}
		}

	}

	return nil
}

// MarshalBinary interface implementation
func (m *BlockLastCommit) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *BlockLastCommit) UnmarshalBinary(b []byte) error {
	var res BlockLastCommit
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}

// BlockLastCommitSignaturesItems0 block last commit signatures items0
//
// swagger:model BlockLastCommitSignaturesItems0
type BlockLastCommitSignaturesItems0 struct {

	// block id flag
	BlockIDFlag int64 `json:"block_id_flag,omitempty"`

	// signature
	// Example: 7uTC74QlknqYWEwg7Vn6M8Om7FuZ0EO4bjvuj6rwH1mTUJrRuMMZvAAqT9VjNgP0RA/TDp6u/92AqrZfXJSpBQ==
	Signature string `json:"signature,omitempty"`

	// timestamp
	// Example: 2017-12-30T05:53:09.287+01:00
	Timestamp string `json:"timestamp,omitempty"`

	// validator address
	ValidatorAddress string `json:"validator_address,omitempty"`
}

// Validate validates this block last commit signatures items0
func (m *BlockLastCommitSignaturesItems0) Validate(formats strfmt.Registry) error {
	return nil
}

// ContextValidate validates this block last commit signatures items0 based on context it is used
func (m *BlockLastCommitSignaturesItems0) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
	return nil
}

// MarshalBinary interface implementation
func (m *BlockLastCommitSignaturesItems0) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *BlockLastCommitSignaturesItems0) UnmarshalBinary(b []byte) error {
	var res BlockLastCommitSignaturesItems0
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
