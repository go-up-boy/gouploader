package gouploader

type Uploader struct {
	storage *Storage
}

func NewUploader(storage *Storage) *Uploader {
	return &Uploader{
		storage: storage,
	}
}