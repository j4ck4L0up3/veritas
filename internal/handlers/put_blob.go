package handlers

import (
	"net/http"
)

func PutBlobStreamHandler() http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
		},
	)
}
