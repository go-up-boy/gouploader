package gouploader

import (
	"errors"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
)

// SizeErr 使用 CheckSeekerMove(hash) 检查文件上传进度；切分后重新上传；
// HashedErr 上传完成后 校验哈希，哈希与前端不一致
// RepeatingErr 上传名称已经存在；重新命名
const (
	SizeErr 	= "upload size error"
	HashedErr     = "file uploaded hash is different"
	RepeatingErr	= "files repeating"
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
type storageStandard interface {
	CheckSeekerMove(hash string) (int64, error)
}

func (uploader *Uploader)SingleUpload(file *multipart.File, header *multipart.FileHeader) singleStandard {
	return &singleUpload{
		File:       *file,
		FileHeader: header,
		storage:    uploader.storage,
	}
}

func (uploader *Uploader)NewStorage() storageStandard {
	return &singleUpload{
		storage:    uploader.storage,
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
	var filename string
	if err := u.initMustParams(); err != nil {
		return "", err
	}
	moveStorage, err := u.storage.Load(hash)
	if err != nil {
		return "", err
	}
	if moveStorage.Empty() {
		filename = u.MoveDir + u.MoveFilename
	} else {
		filename = moveStorage.Filename
	}
	fileInfo, err := os.Stat(filename)
	moveFile, errf := os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_RDWR, os.ModePerm)
	defer u.storeSpeedProgress(&moveStorage)
	defer moveFile.Close()
	if errf != nil {
		return "", errf
	}
	if err == nil {
		if FileMd5(moveFile) == hash {
			moveStorage.Hash = hash
			moveStorage.Size = fileInfo.Size()
			moveStorage.MoveSize = fileInfo.Size()
			moveStorage.Filename = filename
			// hash一致，上传完成
			return moveStorage.Filename, nil
		}
		if moveStorage.Empty() {
			// 当没有hash存储，重名文件 提示 Err
			return moveStorage.Filename, errors.New(RepeatingErr)
		} else {
			// 断点续传文件字节是否符合
			if (fileInfo.Size() + u.FileHeader.Size) != moveStorage.Size {
				return moveStorage.Filename, errors.New(SizeErr)
			}
		}
		moveStorage.Filename = filename
		moveStorage.MoveSize = fileInfo.Size()
	} else if os.IsNotExist(err) {
		moveStorage.Hash = hash
		moveStorage.Filename = u.MoveDir + u.MoveFilename
		moveStorage.MoveSize = 0
		moveStorage.Size = u.FileHeader.Size
	} else {
		return "", err
	}
	_, err = moveFile.Seek(moveStorage.MoveSize, io.SeekStart)
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
		moveStorage.MoveSize = moveStorage.MoveSize + int64(n)
	}
	_, err = moveFile.Seek(0, io.SeekStart)
	if err != nil {
		return "", err
	}
	if hash == FileMd5(moveFile) {
		moveStorage.Hash = hash
		return moveStorage.Filename, nil
	}
	return moveStorage.Filename, errors.New(HashedErr)
}

func (u *singleUpload) CheckSeekerMove(hash string) (int64, error) {
	moveStorage, err := u.storage.Load(hash)
	if err != nil || moveStorage.Empty() {
		return 0, err
	}
	fileInfo, err := os.Stat(moveStorage.Filename)
	if err == nil {
		moveStorage.MoveSize = fileInfo.Size()
	} else if os.IsNotExist(err) {
		return 0, nil
	} else {
		return 0, err
	}
	if err = u.storage.Store(&moveStorage); err != nil {
		return 0, err
	}
	return moveStorage.MoveSize, nil
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
		u.MoveDir = "./uploads"
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
	if file.Hash != "" {
		u.storage.Store(file)
	}
}