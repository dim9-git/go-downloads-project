package entity

type FileMetadata struct {
	ID       string
	MimeType string
	Size     int64
}

type File struct {
	Metadata FileMetadata
	Data     []byte
}
