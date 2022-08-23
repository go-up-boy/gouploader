package gouploader

import (
	"encoding/json"
	"sync"
)

var hashMap sync.Map

type StorageFile struct {
	Filename string
	Hash string
	MoveSize int64
	Size int64
}

type Storage interface {
	Load(hash string) (StorageFile, error)
	Store(file *StorageFile) error
}

type Default struct {
}

func (d *Default) Load(hash string) (StorageFile, error) {
	file := StorageFile{}
	if value, ok := hashMap.Load(hash); ok {
		b, err := json.Marshal(value)
		if err != nil {
			return file, err
		}
		err = json.Unmarshal(b, &file)
		if err != nil {
			return file, err
		}
	}
	return file, nil
}

func (d *Default) Store(file *StorageFile) error {
	hashMap.Store(file.Hash, file)

	return nil
}

func (d StorageFile) Empty() bool {
	return d == StorageFile{}
}

