// Code generated by go-swagger; DO NOT EDIT.

package transactions

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"io"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"

	"github.com/lidofinance/terra-monitors/openapi/models"
)

// GetV1TxsReader is a Reader for the GetV1Txs structure.
type GetV1TxsReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *GetV1TxsReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewGetV1TxsOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	default:
		return nil, runtime.NewAPIError("response status code does not match any response statuses defined for this endpoint in the swagger spec", response, response.Code())
	}
}

// NewGetV1TxsOK creates a GetV1TxsOK with default headers values
func NewGetV1TxsOK() *GetV1TxsOK {
	return &GetV1TxsOK{}
}

/* GetV1TxsOK describes a response with status code 200, with default header values.

Success
*/
type GetV1TxsOK struct {
	Payload *models.GetTxListResult
}

func (o *GetV1TxsOK) Error() string {
	return fmt.Sprintf("[GET /v1/txs][%d] getV1TxsOK  %+v", 200, o.Payload)
}
func (o *GetV1TxsOK) GetPayload() *models.GetTxListResult {
	return o.Payload
}

func (o *GetV1TxsOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.GetTxListResult)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
