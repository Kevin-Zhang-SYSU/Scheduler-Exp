package model

//request.go存放前端发到后端的请求的结构体
//目前是初版结构体，设计的还比较冗余，日后可以优化。

import (
	"time"
)

// 要在其他go文件里访问到这个类，首字母定义要大写
// TaskListInfo是前端请求创建一个容器时，记录下前端传来的信息的类，格式与前端的json对应。
// 注意, K8S对资源的名称有格式要求：只能包含小写字母、数字和- ; 名称必须以字母或数字开头 ; 长度不超253。
// 对名字的检查可以在前端完成也可以在后端完成, 待讨论
type TaskListInfo struct {
	UserName string //这里若变量首字母小写，则是未导出字段，只能在同一个包内被访问和赋值
	TaskName string
	Path     string
}

type LoginInfo struct {
	//记得大写首字母！不然绑定不上去的，gin无法在包外访问结构体字段
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type RegisterInfo struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Repeat   string `json:"repeat" binding:"required"`
}

type TaskSubmitInfo struct {
	TaskId      int
	TaskName    string
	Description string
	Password    string
	Image       string
	Memory      int
	Cpu         int
	Gpu         int
}

type DeploymentModel struct {
	TaskId     int
	DeployInfo string
	Path       string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
