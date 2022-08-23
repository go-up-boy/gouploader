## 1go-Uploader
go-Uploader 表单上传扩展

## 1、Install
    go get github.com/go-up-boy/gouploader@dev

## 2. 快速开始
    // Go-zero 使用
    // 1、定义好结构体，需要生成后手动修改
    type UploadReq struct {
        File       *multipart.File `form:"file,optional"`
        FileHeader *multipart.FileHeader `form:"file_header,optional"`
    }
    // 2、svc 增加 servicecontent
    type ServiceContext struct {
        Config config.Config
        GoUploader *gouploader.Uploader
    }
    func NewServiceContext(c config.Config) *ServiceContext {
        return &ServiceContext{
        Config:     c,
        GoUploader: gouploader.NewUploader(&gouploader.Default{}),
        }
    }
    // 3. handle 接收file
    var req types.UploadReq
    reader, header ,_:= r.FormFile("up_file")
		req.File = &reader
		req.FileHeader = header
    // 4. logic 上传文件
	path, err := l.svcCtx.GoUploader.
		SingleUpload(req.File, req.FileHeader).
		SetMoveDir("./uploads/").
		Move()