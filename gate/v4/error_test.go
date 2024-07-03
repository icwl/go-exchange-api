package gate

import (
	"testing"
)

func TestErrResponse_Error(t *testing.T) {
	var err = NewErrResponse([]byte(`{"message": "test"}`))
	t.Logf("%v", err)
}

func TestErrResponseBody_Error(t *testing.T) {
	var err error = ErrResponseBody([]byte("test body"))
	t.Logf("%v", err)
}
