package coinex

import (
	"fmt"
	"net/http"
)

type ErrResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func NewErrResponse(code int, message string) error {
	return &ErrResponse{
		Code:    code,
		Message: message,
	}
}

func (e *ErrResponse) Error() string {
	return fmt.Sprintf(`{"code":%d,"msg":"%s"}`, e.Code, e.Message)
}

type ErrResponseBody []byte

func (e ErrResponseBody) Error() string {
	return string(e)
}

type ErrResponseStatus struct {
	Code   int
	Status string
}

func NewErrResponseStatus(res *http.Response) error {
	return &ErrResponseStatus{
		Code:   res.StatusCode,
		Status: res.Status,
	}
}

func (e ErrResponseStatus) Error() string {
	return fmt.Sprintf("%d: %s", e.Code, e.Status)
}
