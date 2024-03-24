package main

import (
	"fmt"
	"math/rand"
	"scheduler-exp/service"
	"scheduler-exp/util"
	"time"
)

// Task 结构体
type Task struct {
	taskName string
	gpumem   int
	exist    bool
}

func generateTask(index int, taskChan chan<- Task) {
	source := rand.NewSource(time.Now().UnixNano())
	randGenerator := rand.New(source)
	gpumem := randGenerator.Intn(6) + 1
	taskName := fmt.Sprintf("task-%d-%d", index, gpumem)
	task := Task{
		taskName: taskName,
		gpumem:   gpumem,
		exist:    false,
	}
	taskChan <- task
}

func createTask(taskList []Task) {
	// 循环taskNum次，每次从taskMap中取出一个任务，每次相隔5秒
	for i := 0; i < len(taskList); i++ {
		time.Sleep(5 * time.Second)
		taskName := taskList[i].taskName
		gpumem := taskList[i].gpumem
		util.Logger.Debugf("taskName: %s, gpumem: %d", taskName, gpumem)
		// 创建新的taskSubmitInfo
		taskSubmitInfo := service.Create_taskSubmitInfo(taskName, gpumem)
		// 将submitTask传入CreateTask函数中
		var userName string = "exp"
		create, err := service.CreateTask(userName, *taskSubmitInfo)
		if err != nil {
			util.Logger.Errorf("Failed to create task %s", taskName)
		}
		if create {
			util.Logger.Infof("Task %s created successfully", taskName)
		}
		// 检查任务是否存在
		exist, err := service.CheckTask(userName, *taskSubmitInfo)
		if err != nil {
			util.Logger.Errorf("Failed to check task %s", taskName)
		}
		if exist {
			util.Logger.Infof("Task %s exists", taskName)
		}
	}
}

func main() {
	fmt.Println("This is the scheduler experiment system")
	var taskNum int = 10
	taskList := make([]Task, taskNum)

	// 创建一个通道来收集生成的任务
	taskChan := make(chan Task, taskNum)

	// 并发生成任务列表
	for i := 0; i < taskNum; i++ {
		go generateTask(i, taskChan)
	}

	// 从通道中收集生成的任务
	for i := 0; i < taskNum; i++ {
		task := <-taskChan
		taskList[i] = task
	}

	// 打印生成的任务列表
	for i, task := range taskList {
		fmt.Printf("Task %d: %+v\n", i+1, task)
	}

	// 创建任务,传入taskList
	// createTask(taskList)

}
