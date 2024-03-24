// 这个文件是关于后端部分，接受前端用户请求并创建deployment和service的代码
package service

import (
	"scheduler-exp/k8s"
	"scheduler-exp/model"
	"scheduler-exp/util"
)

var taskSubmitInfo = model.TaskSubmitInfo{
	TaskId:      0,
	TaskName:    "default",
	Description: "default",
	Password:    "ndc@wwg1234",
	Image:       "default",
	Memory:      0,
	Cpu:         0,
	Gpu:         0,
}

var userId string
var userName string
var TaskNameHasUsed bool = false

func Create_taskSubmitInfo(taskName string, gpumem int) *model.TaskSubmitInfo {
	return &model.TaskSubmitInfo{
		TaskId:      0,
		TaskName:    taskName,
		Description: "default",
		Password:    "ndc@wwg1234",
		Image:       "default",
		Memory:      0,
		Cpu:         0,
		Gpu:         gpumem,
	}
}

func CreateTask(userName string, taskSubmitInfo model.TaskSubmitInfo) (bool, error) {
	// // 将信息通过日志输出到控制台
	util.Logger.Info("Create task ", taskSubmitInfo.TaskName, "with gpu ", taskSubmitInfo.Gpu)
	if !CheckResource(taskSubmitInfo.Gpu) {
		util.Logger.Error("Insufficient resources")
		return false, nil
	}
	err := k8s.CreateDeployment(userName, taskSubmitInfo)
	if err != nil {
		util.Logger.Error("Failed to create deployment")
		return false, err
	}
	return true, nil
}

// 查找Task是否已经存在
func CheckTask(userName string, taskName string) (bool, error) {
	// 查找Task是否已经存在
	if k8s.GetDeployment(userName+"-"+taskName) != nil {
		return true, nil
	}
	return false, nil
}

// 删除Task
func DeleteTask(userName, taskName string) error {
	// 删除Task
	err := k8s.DeleteDeployment(userName + "-" + taskName)
	if err != nil {
		util.Logger.Error("Failed to delete deployment")
		return err
	}
	return nil
}
