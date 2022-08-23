package gouploader

import (
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
)

type singleUpload struct {
	File       multipart.File
	FileHeader *multipart.FileHeader
	MoveDir string
	MoveFilename string
	LimitExt []string
	storage Storage
}

type singleStandard interface {
	Move() (string, error)
	SetMoveFilename(filename string) *singleUpload
	SetMoveDir(dir string) *singleUpload
	SetAllowExt(ext []string) *singleUpload
}

func (uploader *Uploader)SingleUpload(file *multipart.File, header *multipart.FileHeader) singleStandard {
	return &singleUpload{
		File: *file,
		FileHeader: header,
		storage: *uploader.storage,
	}
}

func (u *singleUpload) Move() (string, error) {
	defer u.File.Close()
	if err := u.initMustParams(); err != nil {
		return "", err
	}
	filename := u.MoveDir + u.MoveFilename
	moveFile, err := os.OpenFile(filename, os.O_CREATE|os.O_RDWR, os.ModePerm)
	defer moveFile.Close()
	if err != nil {
		return "", err
	}
	buffer := make([]byte, 32768)
	for {
		n, err := u.File.Read(buffer)
		if err != nil {
			if err == io.EOF{
				break
			}
		}
		_, err = moveFile.Write(buffer[:n])
		if err != nil {
			return "", err
		}
	}
	return filename, nil
}

func (u *singleUpload) SeekerMove(hash string) (string, error) {
	defer u.File.Close()
	if err := u.initMustParams(); err != nil {
		return "", err
	}
	moveStorage, err := u.storage.Load(hash)
	if moveStorage.Empty() {
		moveStorage.Hash = hash
		moveStorage.Filename = u.MoveDir + u.MoveFilename
		moveStorage.Size = u.FileHeader.Size
	}
	fileInfo, err := os.Stat(moveStorage.Filename)
	if err == nil {
		if fileInfo.Size() == u.FileHeader.Size {
			return moveStorage.Filename, nil
		}
		moveStorage.MoveSize = fileInfo.Size()
		_, err = u.File.Seek(fileInfo.Size(), io.SeekStart)
		if err != nil {
			return "", err
		}
	} else if os.IsNotExist(err) {
		moveStorage.Hash = hash
		moveStorage.Filename = u.MoveDir + u.MoveFilename
		moveStorage.Size = u.FileHeader.Size
	}
	moveFile, err := os.OpenFile(moveStorage.Filename, os.O_CREATE|os.O_APPEND|os.O_RDWR, os.ModePerm)
	defer moveFile.Close()
	if err != nil {
		return "", err
	}
	buffer := make([]byte, 32768)
	for {
		n, err := u.File.Read(buffer)
		if err != nil {
			if err == io.EOF{
				break
			}
		}
		_, err = moveFile.Write(buffer[:n])
		moveStorage.MoveSize = moveStorage.MoveSize + int64(n)
		if err != nil {
			return "", err
		}
	}
	u.storeSpeedProgress(&moveStorage)
	return moveStorage.Filename, nil
}

func (u *singleUpload) initMustParams() error {
	_, err := CheckExtName(u.FileHeader.Filename, u.LimitExt)
	if err != nil {
		return err
	}
	if u.MoveFilename == "" {
		u.MoveFilename = u.FileHeader.Filename
	}
	if u.MoveDir == "" {
		u.MoveDir = "/uploads"
	}
	if _, err := os.Stat(u.MoveDir); err != nil {
		if os.IsNotExist(err) {
			os.MkdirAll(u.MoveDir, os.ModeDir)
		} else {
			return err
		}
	}
	return nil
}

func (u *singleUpload) SetMoveFilename(filename string) *singleUpload {
	u.MoveFilename = filename
	return u
}

func (u *singleUpload) SetMoveDir(dir string) *singleUpload {
	u.MoveDir = strings.TrimRight(filepath.ToSlash(dir), "/") + "/"
	return u
}

func (u *singleUpload) SetAllowExt(ext []string) *singleUpload {
	u.LimitExt = ext
	return u
}

func (u *singleUpload) storeSpeedProgress(file *StorageFile) {
	u.storage.Store(file)
}