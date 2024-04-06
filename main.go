package main

import (
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"scheduler-exp/service"
	"scheduler-exp/util"
	"syscall"
	"time"
)

// Task 结构体
type Task struct {
	taskName string
	gpumem   int
	exist    bool
	node     string
	info     string
}

var userName string = "exp"

func generateTask(index int) Task {
	seed := int64(index)
	source := rand.NewSource(seed)
	randGenerator := rand.New(source)
	// source := rand.NewSource(time.Now().UnixNano())
	// randGenerator := rand.New(source)
	gpumem := randGenerator.Intn(6) + 1
	taskName := fmt.Sprintf("task-%d-%d", index, gpumem)
	task := Task{
		taskName: taskName,
		gpumem:   gpumem,
		exist:    false,
		node:     "",
		info:     "",
	}
	return task
}

func createTask(taskList []Task, stop <-chan struct{}) {
	// 循环taskNum次，每次从taskMap中取出一个任务，每次相隔5秒
	for i := 0; i < len(taskList); i++ {
		select {
		case <-stop:
			fmt.Println("Stopping task creation...")
			return
		default:
			time.Sleep(3 * time.Second)
			taskName := taskList[i].taskName
			gpumem := taskList[i].gpumem
			util.Logger.Debugf("taskName: %s, gpumem: %d", taskName, gpumem)
			// 创建新的taskSubmitInfo
			taskSubmitInfo := service.Create_taskSubmitInfo(taskName, gpumem)
			// 将submitTask传入CreateTask函数中
			create, err := service.CreateTask(userName, *taskSubmitInfo)
			if err != nil {
				util.Logger.Errorf("Failed to create task %s", taskName)
			}
			if create {
				util.Logger.Infof("Task %s created successfully", taskName)
			} else {
				taskList[i].info = "Insufficient resources"
			}
			// 检查任务是否存在
			nodeName, err := service.CheckTask(userName, taskName)
			if err != nil {
				util.Logger.Errorf("Failed to check task %s", taskName)
			}
			if nodeName != "" {
				util.Logger.Infof("Task %s exists", taskName)
				taskList[i].exist = true
				taskList[i].node = nodeName
			}
		}
	}
}

// 删除任务
func deleteTask(taskList []Task) {
	for i := 0; i < len(taskList); i++ {
		if !taskList[i].exist {
			continue
		}
		time.Sleep(3 * time.Second)
		taskName := taskList[i].taskName
		err := service.DeleteTask("exp", taskName)
		if err != nil {
			util.Logger.Errorf("Failed to delete task %s", taskName)
		}
		// 检查任务是否存在
		nodeName, err := service.CheckTask(userName, taskName)
		if err != nil {
			util.Logger.Errorf("Failed to check task %s", taskName)
		}
		if nodeName == "" {
			util.Logger.Infof("Task %s exists", taskName)
			taskList[i].exist = false
		}
	}
}

func main() {
	fmt.Println("This is the scheduler experiment system")
	var taskNum int = 30
	taskList := make([]Task, taskNum)

	// 生成任务列表
	for i := 0; i < taskNum; i++ {
		taskList[i] = generateTask(i)
	}

	// 打印生成的任务列表
	for i, task := range taskList {
		fmt.Printf("Task %d: %+v\n", i, task)
	}

	// 创建任务,传入taskList
	fmt.Println("Begin create task ... ... ...")

	// 通道用于通知创建任务停止
	stop := make(chan struct{})

	// 监听中断信号
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	// 创建任务的 goroutine
	done := make(chan struct{})
	go func() {
		createTask(taskList, stop)
		close(done)
	}()

	// 等待中断信号或创建任务完成
	select {
	case <-c:
		fmt.Println("\nReceived interrupt signal. Deleting tasks...")
		close(stop) // 向 createTask 发送停止信号
		<-done      // 等待 createTask 完成
	case <-done:
		fmt.Println("Task creation completed.")
	}

	// 打印任务信息
	fmt.Println("Task information after create:")
	for i, task := range taskList {
		fmt.Printf("Task %d: %+v\n", i, task)
	}
	// 程序阻塞，要求用户输入任意字符
	fmt.Println("Press Enter to delete tasks...")
	fmt.Scanln()

	// 删除任务
	fmt.Println("Begin delete task ... ... ...")
	deleteTask(taskList)

	// 打印任务信息
	fmt.Println("Task information after delete:")
	for i, task := range taskList {
		fmt.Printf("Task %d: %+v\n", i, task)
	}
}
