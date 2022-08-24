## 1go-Uploader
go-Uploader 表单上传扩展，支持秒传、断点续传、自定义存储Hash

## 1、Install
    go get github.com/go-up-boy/gouploader

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


## 3. 断点续传、秒传的使用
    // 前端计算文件 md5 32位 hash值
    // 1、使用该方法 检查文件上传进度，返回已经上传字节数
    l.svcCtx.GoUploader.
    NewStorage().
    CheckSeekerMove("bb53183243f4485383e1ea4bdf1e954a")

    // 2、上传文件，记得 前端根据 方法CheckSeekerMove(hash) 返回字节数 切分文件
    path, err := l.svcCtx.GoUploader.
		SingleUpload(req.File, req.FileHeader).
		SetMoveDir("./uploads/").
		SeekerMove("ca3e4b36c225c370baa3062f347386de")
## 4. Storage存储文件哈希接口
    // 实现 Storage 接口,传入初始化参数即可
    gouploader.NewUploader(&gouploader.Default{})

    type Storage interface {
        Load(hash string) (StorageFile, error)
        Store(file *StorageFile) error
    }
    // 注意: 一定要包含结构体字段
    type StorageFile struct {
        Filename string
        Hash string
        MoveSize int64
        Size int64
    }

## 5、流程图
![post](https://s1.ax1x.com/2022/08/24/vgptl8.png)