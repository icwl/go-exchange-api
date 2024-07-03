package gate

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type ErrResponse struct {
	Label   string
	Message string
}

func NewErrResponse(body []byte) error {
	e := new(ErrResponse)
	if err := json.Unmarshal(body, &e); err != nil {
		err := ErrResponseBody(body)
		return err
	}
	return e
}

func (e *ErrResponse) Error() string {
	return fmt.Sprintf("label:%s msg:%s", e.Label, e.Message)
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
