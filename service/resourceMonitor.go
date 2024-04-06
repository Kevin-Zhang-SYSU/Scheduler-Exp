package service

import (
	"os/exec"
	"scheduler-exp/model"
	"scheduler-exp/util"
	"strconv"
	"strings"
)

// 解析kubectl inspect gpushare命令的字符串输入
func parseGPUInfo(input string) ([]model.GPUInfo, error) {
	var gpuInfoList []model.GPUInfo
	lines := strings.Split(input, "\n")

	for _, line := range lines[1 : len(lines)-4] { //跳过输入的第一行，第一行是表头信息。最后一行是空行，倒数2-4行是总结信息，都跳过
		fields := strings.Fields(line) //以空白字符?为划分获取该行的每一个词元素

		// 如果字段数量小于 7，则将缺失的部分填充为 "0/0"
		gpuFields := make([]string, 7)
		if len(fields) == 7 {
			copy(gpuFields, fields)
		} else {
			for i := range gpuFields {
				gpuFields[i] = "0/0"
			}

			if len(fields) < 7 {
				for i := 0; i < len(fields)-1; i++ {
					gpuFields[i] = fields[i]
				}
				gpuFields[6] = fields[len(fields)-1]
			}
		}

		gpuInfo := model.GPUInfo{
			Name:         gpuFields[0],
			IPAddress:    gpuFields[1],
			GPU0:         gpuFields[2],
			GPU1:         gpuFields[3],
			GPU2:         gpuFields[4],
			GPU3:         gpuFields[5],
			GPUMemoryGiB: gpuFields[6],
		}

		gpuInfoList = append(gpuInfoList, gpuInfo)
	}

	return gpuInfoList, nil
}

// 终端执行kubectl inspect gpushare
func GetGPU() []model.GPUInfo {
	cmd := exec.Command("kubectl", "inspect", "gpushare")
	output, err := cmd.CombinedOutput()
	if err != nil {
		util.Logger.Error("Failed to execute kubectl inspect gpushare")
		return nil
	}

	gpuInfoList, err := parseGPUInfo(string(output))
	if err != nil {
		util.Logger.Error("Failed to parse GPU info")
	} else {
		util.Logger.Info("Successfully get GPU info")
	}
	// 打印gpuInfoList的信息
	for _, gpuInfo := range gpuInfoList {
		util.Logger.Debug(gpuInfo.Name, "has gpus: ", gpuInfo.GPU0, gpuInfo.GPU1, gpuInfo.GPU2, gpuInfo.GPU3)
	}
	return gpuInfoList
}

// 计算集群的GPU内存利用率
func Com_GPU_MEM() {
	gpuInfoList := GetGPU()
	for _, gpuInfo := range gpuInfoList {
		used, total := strings.Split(gpuInfo.GPUMemoryGiB, "/")[0], strings.Split(gpuInfo.GPUMemoryGiB, "/")[1]
		usedNum, err := strconv.Atoi(used)
		totalNum, err := strconv.Atoi(total)
	}
	return
}

// 检查集群资源是否满足要求
func CheckResource(GpuRequest int) bool {
	// 检查GPU资源是否满足要求
	cmd := exec.Command("kubectl", "inspect", "gpushare")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return false
	}
	gpuInfoList, err := parseGPUInfo(string(output))
	if err != nil {
		return false
	}
	for _, gpuInfo := range gpuInfoList {
		util.Logger.Debugln(gpuInfo.Name, "has gpus:", gpuInfo.GPU0, gpuInfo.GPU1, gpuInfo.GPU2, gpuInfo.GPU3)
		for _, gpu := range []string{gpuInfo.GPU0, gpuInfo.GPU1, gpuInfo.GPU2, gpuInfo.GPU3} {
			if gpu == "0/0" {
				continue
			}
			used, total := strings.Split(gpu, "/")[0], strings.Split(gpu, "/")[1]
			usedNum, err := strconv.Atoi(used)
			if err != nil {
				return false
			}
			totalNum, err := strconv.Atoi(total)
			if err != nil {
				return false
			}
			if totalNum-usedNum >= GpuRequest {
				util.Logger.Debug("GPU request is satisfied")
				return true
			}
		}
	}
	return false
}
