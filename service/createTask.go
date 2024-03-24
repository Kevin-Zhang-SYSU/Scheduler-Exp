// 这个文件是关于后端部分，接受前端用户请求并创建deployment和service的代码
package service

import (
	"fmt"
	"goClient/server/exceptions"
	"goClient/server/k8s"
	"goClient/server/middleware"
	"goClient/server/model"
	"math/rand"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

var jupyterlab_port = 31000
var codeserver_port = 31001

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

func BindTaskSubmitInfo(context *gin.Context) {
	middleware.Logger.Info("Bind Task Submit Info")
	userId = context.Request.Header.Get("userId")
	userName = context.Request.Header.Get("userName")
	// 如果绑定失败，则返回错误响应并终止处理
	if err := context.ShouldBind(&taskSubmitInfo); err != nil {
		myErr := exceptions.VALID_ERROR
		myErr.Data = err.Error()
		context.Error(myErr)
		context.Abort()
		return
	}
	// 检查集群资源是否满足要求，如果不满足则返回错误响应并终止处理，不满足返回false
	if !CheckResource(taskSubmitInfo.Gpu) {
		myErr := exceptions.INSUFFICIENT_RESOURCES
		myErr.Data = myErr.Msg
		context.Error(myErr)
		context.Abort()
		return
	}
	middleware.Logger.Debug("Bind Task Submit Info success")
}

// 检查任务名是否重复。已优化：实际情况里不同用户的任务名应该是可以重复的
func checkTaskNameUsed(userId string, taskName string) (bool, error) {
	middleware.Logger.Debug("check task name used")
	var tasks []model.Task
	result := model.DB.Where("user_id = ? AND task_name = ?", userId, taskName).Find(&tasks)
	if result.Error != nil {
		myErr := exceptions.DATABASE_ERROR
		myErr.Data = result.Error.Error()
		return true, myErr
	}
	if len(tasks) >= 1 {
		return true, nil
	}
	return false, nil
}

// 检查端口是否已被使用,是就返回true,不是就false
func checkPortUsed(port int) (bool, error) {
	middleware.Logger.Debug("check pot used")
	var tasks []model.Task
	result := model.DB.Where("jupyterlab_port = ? OR codeserver_port = ?", port, port).Find(&tasks)
	if result.Error != nil {
		myErr := exceptions.DATABASE_ERROR
		myErr.Data = result.Error.Error()
		return true, myErr //有错误就返回true
	}
	if len(tasks) >= 1 {
		return true, nil
	}
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return true, nil
	}
	defer listener.Close()
	return false, nil
}

func getAvailablePort() (int, error) {
	source := rand.NewSource(time.Now().UnixNano())
	randGenerator := rand.New(source)
	for {
		port := randGenerator.Intn(32760-30000) + 30000
		isUsed, err := checkPortUsed(port)
		if err != nil {
			return 0, err
		}
		if !isUsed {
			return port, nil
		}
	}
}

func getAvailablePortExcept(excludePort int) (int, error) {
	source := rand.NewSource(time.Now().UnixNano())
	randGenerator := rand.New(source)
	for {
		port := randGenerator.Intn(32760-30000) + 30000
		isUsed, err := checkPortUsed(port)
		if port != excludePort && !isUsed && err == nil {
			return port, nil
		}
	}
}

func CreateTask(context *gin.Context) {
	// // 将信息通过日志输出到控制台
	middleware.Logger.Info("Create task request from userId: ", userId+" userName: "+userName)
	isUsed, err := checkTaskNameUsed(userId, taskSubmitInfo.TaskName)
	if err != nil {
		context.Error(err)
		return
	}
	if isUsed {
		TaskNameHasUsed = true
		middleware.Logger.Debug("taskname has benn used return error!")
		myErr := exceptions.DUPLICATE_TASKNAME
		myErr.Data = myErr.Msg
		context.Error(myErr)
		return
	}
	err = k8s.CreateDeployment(userName, taskSubmitInfo)
	if err != nil {
		context.Error(err)
		return
	}

	jupyterlab_port, err = getAvailablePort()
	if err != nil {
		context.Error(err)
		return
	}

	codeserver_port, err = getAvailablePortExcept(jupyterlab_port)
	if err != nil {
		context.Error(err)
		return
	}

	err = k8s.CreateService(userName, taskSubmitInfo, jupyterlab_port, codeserver_port)
	if err != nil {
		context.Error(err)
		return
	}
}

// 把创建的任务记录到数据库中
// 这个过程中出错就返回false，没错返回true
func recordTask(context *gin.Context, ip string) bool {
	userId_int, err := strconv.Atoi(userId)
	if err != nil {
		myErr := exceptions.TYPE_ERROR
		myErr.Data = err.Error()
		context.Error(myErr)
		return false
	}
	task := model.Task{
		// TaskId:    taskSubmitInfo.TaskId,
		UserId:         userId_int,
		TaskName:       taskSubmitInfo.TaskName,
		Descri:         taskSubmitInfo.Description,
		Image:          taskSubmitInfo.Image,
		Cpu:            taskSubmitInfo.Cpu,
		Memory:         taskSubmitInfo.Memory,
		Passwd:         taskSubmitInfo.Password,
		Gpu:            taskSubmitInfo.Gpu,
		Ip:             ip,
		JupyterlabPort: jupyterlab_port,
		CodeserverPort: codeserver_port,
		UpdatedAt:      time.Now(),
		CreatedAt:      time.Now(),
	}
	result := model.DB.Create(&task)
	if result.Error != nil {
		myErr := exceptions.DATABASE_ERROR
		myErr.Data = result.Error.Error()
		context.Error(myErr)
		//如果插入数据库失败，要把之前创建的deployment和service删除了。
		k8s.DeleteDeployment(userName + "-" + task.TaskName)
		k8s.DeleteService(userName + "-" + task.TaskName)
		return false
	}
	return true
}

func Response(context *gin.Context) {
	middleware.Logger.Debug("creakTask Response")
	// 如果已经判断任务名被占用，直接返回，以免删除之前创建的任务
	if TaskNameHasUsed {
		TaskNameHasUsed = false
		return
	}
	// 凡是传参为taskSubmitInfo.TaskName都要注意，要加上userName判断
	ip, err := k8s.GetNodeIPOfDeploymentPod(userName + "-" + taskSubmitInfo.TaskName)
	middleware.Logger.Debug("Get node ip:", ip)
	if err != nil {
		middleware.Logger.Debug("Get node ip of deploy pod error!", ip)
		context.Error(err)
		k8s.DeleteDeployment(userName + "-" + taskSubmitInfo.TaskName)
		k8s.DeleteService(userName + "-" + taskSubmitInfo.TaskName)
		return
	}
	if !recordTask(context, ip) {
		middleware.Logger.Debug("Fail to record task!")
		return
	}
	middleware.Logger.Debug("Start to make reply")
	reply := model.SubmitTaskReply{}
	reply.JupyterAddress = ip + ":" + strconv.Itoa(jupyterlab_port)
	reply.CodeServerAddress = ip + ":" + strconv.Itoa(codeserver_port)
	reply.Message = userName + ", your " + taskSubmitInfo.TaskName + " created successfully"
	context.JSON(http.StatusOK, reply)
}
