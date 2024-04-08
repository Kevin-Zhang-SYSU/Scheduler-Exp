// 这个文件是关于后端部分，接受前端用户请求并创建deployment和service的代码
package service

import (
	"context"
	"fmt"
	"scheduler-exp/k8s"
	"scheduler-exp/model"
	"scheduler-exp/util"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	util.Logger.Infoln("Create task ", taskSubmitInfo.TaskName, " with gpu ", taskSubmitInfo.Gpu)
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
func CheckTask(userName string, taskName string) (string, error) {
	// 查找Task是否已经存在
	time.Sleep(3 * time.Second)
	deploymentName := userName + "-" + taskName
	deploy := k8s.GetDeployment(deploymentName)
	if deploy != nil {
		// 找到deploy对应的pod
		// 获取指定名称的Deployment的所有Pod
		pods, err := k8s.ClientSet.CoreV1().Pods("").List(context.Background(), metav1.ListOptions{
			LabelSelector: fmt.Sprintf("task=%s", deploymentName),
		})
		if err != nil {
			panic(err.Error())
		}
		pod := pods.Items[0]
		nodeName := pod.Spec.NodeName
		util.Logger.Info("Task exists on node ", nodeName)
		return nodeName, nil
	}
	util.Logger.Error("Deploy not found")
	return "", nil
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
