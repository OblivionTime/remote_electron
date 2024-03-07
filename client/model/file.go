package model

type FileInfo struct {
	FileName   string `json:"file_name,omitempty"`
	FilePath   string `json:"file_path,omitempty"`
	FileSize   int64  `json:"file_size,omitempty"`
	IsDir      bool   `json:"is_dir,omitempty"`
	UpdateTime string `json:"update_time,omitempty"`
}

// 前端发送的数据
type ClientFileReceive struct {
	Op string `json:"op,omitempty"`
	//文件上传和下载
	FileOp      string `json:"fileop,omitempty"`
	FilePath    string `json:"filepath,omitempty"`
	FileSize    int64  `json:"filesize,omitempty"`
	FileContent []byte `json:"filecontent,omitempty"`
	Offset      int    `json:"offset,omitempty"`
	//获取本地目录下所有文件和远程目录所有文件
	LocalPath  string `json:"localpath,omitempty"`
	RemotePath string `json:"remotepath,omitempty"`
	//发送邀请
	Device     string `json:"device,omitempty"`
	SendDevice string `json:"send_device,omitempty"`
	Code       string `json:"code"`

	//下载相关参数
	Speed    string `json:"speed,omitempty"`
	Progress int    `json:"progress,omitempty"`
	Status   int    `json:"status,omitempty"` //0表示开始 1表示正在进行 2表示已完成 3表示出现错误
}
type ClientFileResponse struct {
	Op              string            `json:"op,omitempty"`
	ErrMsg          string            `json:"errmsg,omitempty"`
	LocalFileList   []FileInfo        `json:"local_file_list,omitempty"`
	LocalBasicList  map[string]string `json:"local_basic_list,omitempty"`
	RemoteFileList  []FileInfo        `json:"remote_file_list,omitempty"`
	RemoteBasicList map[string]string `json:"remote_basic_list,omitempty"`

	//文件上传相关参数
	FilePath   string `json:"filepath,omitempty"`
	FileSize   string `json:"filesize,omitempty"`
	FileName   string `json:"filename,omitempty"`
	RemotePath string `json:"remotepath,omitempty"`
	Speed      string `json:"speed,omitempty"`
	Progress   int    `json:"progress,omitempty"`
	Status     int    `json:"status,omitempty"` //0表示开始 1表示正在进行 2表示已完成 3表示出现错误
}
