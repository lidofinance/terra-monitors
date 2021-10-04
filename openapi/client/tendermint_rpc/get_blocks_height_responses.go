// Code generated by go-swagger; DO NOT EDIT.

package tendermint_rpc

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"io"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"

	"github.com/lidofinance/terra-monitors/openapi/models"
)

// GetBlocksHeightReader is a Reader for the GetBlocksHeight structure.
type GetBlocksHeightReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *GetBlocksHeightReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewGetBlocksHeightOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	case 400:
		result := NewGetBlocksHeightBadRequest()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 404:
		result := NewGetBlocksHeightNotFound()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 500:
		result := NewGetBlocksHeightInternalServerError()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	default:
		return nil, runtime.NewAPIError("response status code does not match any response statuses defined for this endpoint in the swagger spec", response, response.Code())
	}
}

// NewGetBlocksHeightOK creates a GetBlocksHeightOK with default headers values
func NewGetBlocksHeightOK() *GetBlocksHeightOK {
	return &GetBlocksHeightOK{}
}

/* GetBlocksHeightOK describes a response with status code 200, with default header values.

The block at a specific height
*/
type GetBlocksHeightOK struct {
	Payload *models.BlockQuery
}

func (o *GetBlocksHeightOK) Error() string {
	return fmt.Sprintf("[GET /blocks/{height}][%d] getBlocksHeightOK  %+v", 200, o.Payload)
}
func (o *GetBlocksHeightOK) GetPayload() *models.BlockQuery {
	return o.Payload
}

func (o *GetBlocksHeightOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.BlockQuery)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewGetBlocksHeightBadRequest creates a GetBlocksHeightBadRequest with default headers values
func NewGetBlocksHeightBadRequest() *GetBlocksHeightBadRequest {
	return &GetBlocksHeightBadRequest{}
}

/* GetBlocksHeightBadRequest describes a response with status code 400, with default header values.

Invalid height
*/
type GetBlocksHeightBadRequest struct {
}

func (o *GetBlocksHeightBadRequest) Error() string {
	return fmt.Sprintf("[GET /blocks/{height}][%d] getBlocksHeightBadRequest ", 400)
}

func (o *GetBlocksHeightBadRequest) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	return nil
}

// NewGetBlocksHeightNotFound creates a GetBlocksHeightNotFound with default headers values
func NewGetBlocksHeightNotFound() *GetBlocksHeightNotFound {
	return &GetBlocksHeightNotFound{}
}

/* GetBlocksHeightNotFound describes a response with status code 404, with default header values.

Request block height doesn't
*/
type GetBlocksHeightNotFound struct {
}

func (o *GetBlocksHeightNotFound) Error() string {
	return fmt.Sprintf("[GET /blocks/{height}][%d] getBlocksHeightNotFound ", 404)
}

func (o *GetBlocksHeightNotFound) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	return nil
}

// NewGetBlocksHeightInternalServerError creates a GetBlocksHeightInternalServerError with default headers values
func NewGetBlocksHeightInternalServerError() *GetBlocksHeightInternalServerError {
	return &GetBlocksHeightInternalServerError{}
}

/* GetBlocksHeightInternalServerError describes a response with status code 500, with default header values.

Server internal error
*/
type GetBlocksHeightInternalServerError struct {
}

func (o *GetBlocksHeightInternalServerError) Error() string {
	return fmt.Sprintf("[GET /blocks/{height}][%d] getBlocksHeightInternalServerError ", 500)
}

func (o *GetBlocksHeightInternalServerError) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	return nil
}
