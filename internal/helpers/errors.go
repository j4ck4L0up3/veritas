package helpers

import (
	"encoding/json/v2"
)

type Code string

const (
	BLOB_UNKNOWN          Code = "BLOB_UNKNOWN"
	BLOB_UPLOAD_INVALID   Code = "BLOB_UPLOAD_INVALID"
	BLOB_UPLOAD_UNKNOWN   Code = "BLOB_UPLOAD_UNKNOWN"
	DIGEST_INVALID        Code = "DIGEST_INVALID"
	MANIFEST_UNKNOWN      Code = "MANIFEST_UNKNOWN"
	MANIFEST_INVALID      Code = "MANIFEST_INVALID"
	MANIFEST_BLOB_UNKNOWN Code = "MANIFEST_BLOB_UNKNOWN"
	UNSUPPORTED           Code = "UNSUPPORTED"
)

var messages = map[Code]string{
	BLOB_UNKNOWN:          "blob unknown to registry",
	BLOB_UPLOAD_INVALID:   "blob upload invalid",
	BLOB_UPLOAD_UNKNOWN:   "blob upload unknown to registry",
	DIGEST_INVALID:        "provided digest did not match content",
	MANIFEST_UNKNOWN:      "manifest unknown",
	MANIFEST_INVALID:      "manifest invalid",
	MANIFEST_BLOB_UNKNOWN: "unknown blob in manifest",
	UNSUPPORTED:           "the operation is unsupported",
}

type Error struct {
	Code    Code   `json:"code"`
	Message string `json:"message;omitempty"`
	Detail  string `json:"detail;omitempty"`
}

type errorBody struct {
	Errors []*Error `json:"errors"`
}

func NewError(code Code, message, detail string) *Error {
	return &Error{
		Code:    code,
		Message: messages[code],
		Detail:  detail,
	}
}

func (e *Error) GetJSON() ([]byte, error) {
	body, err := json.Marshal(errorBody{Errors: []*Error{e}})
	if err != nil {
		return nil, err
	}

	return body, nil
}
