package mw

import (
	"context"
	"gin-quickstart/pkg/reqmeta"
	"net/http"

	"github.com/google/uuid"
)

const (
	HeaderXRequestID = "X-Request-ID"
)

type requestMetadataKey struct{}

func RequestMetadata(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get(HeaderXRequestID)
		if requestID == "" {
			requestID = uuid.New().String()
		}
		w.Header().Set(HeaderXRequestID, requestID)

		method := r.Method
		url := r.URL.String()

		metadata := reqmeta.NewRequestMetadata(requestID, reqmeta.WithHTTPMetadata(reqmeta.HTTPMetadata{
			Method: &method,
			URL:    &url,
		}))

		ctx := r.Context()
		ctx = context.WithValue(ctx, requestMetadataKey{}, metadata)

		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}
