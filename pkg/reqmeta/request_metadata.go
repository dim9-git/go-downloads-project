package reqmeta

type RequestMetadata struct {
	RequestID string
	HTTPMetadata
}

type HTTPMetadata struct {
	Method *string
	URL    *string
}

type Option func(*RequestMetadata)

func NewRequestMetadata(requestID string, options ...Option) *RequestMetadata {
	rm := &RequestMetadata{
		RequestID: requestID,
	}

	for _, option := range options {
		option(rm)
	}

	return rm
}

func WithHTTPMetadata(hmd HTTPMetadata) Option {
	return func(rm *RequestMetadata) {
		rm.HTTPMetadata = hmd
	}
}
