package k8s

import (
	"context"
	"scheduler-exp/model"
	"scheduler-exp/util"

	coreV1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// 创建Service将容器端口对外开放
func CreateService(userName string, taskSubmitInfo model.TaskSubmitInfo, jupyterlab_port int, codeserver_port int) error {
	namespace := "default"
	service := &coreV1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: userName + "-" + taskSubmitInfo.TaskName,
		},
		Spec: coreV1.ServiceSpec{
			Selector: map[string]string{
				"task": userName + "-" + taskSubmitInfo.TaskName,
			},
			Type: coreV1.ServiceTypeNodePort,
			Ports: []coreV1.ServicePort{
				{
					Name:       "jupyterlab-port",
					Port:       8080,
					TargetPort: intstr.FromInt(8888),
					NodePort:   int32(jupyterlab_port),
				},
				{
					Name:       "codeserver-port",
					Port:       8081,
					TargetPort: intstr.FromInt(8889),
					NodePort:   int32(codeserver_port),
				},
			},
		},
	}
	util.Logger.Debug("Creating service ... ... ...")
	result, err := ClientSet.CoreV1().Services(namespace).Create(context.TODO(), service, metav1.CreateOptions{})
	if err != nil {
		util.Logger.Error("Failed to create service %q: %v", result.GetObjectMeta().GetName(), err)
		return err
	}
	util.Logger.Debug("Created service %q.", result.GetObjectMeta().GetName())
	return nil
}

// 删除一个Service，要满足指定的名字和labelSelector="owner=userName"
func DeleteService(serviceName string) error {
	namespace := "default"
	err := ClientSet.CoreV1().Services(namespace).Delete(context.TODO(), serviceName, metav1.DeleteOptions{})
	if err != nil {
		util.Logger.Error("Failed to delete service %q: %v", serviceName, err)
		return err
	} else {
		util.Logger.Debug("Deleted service %q.", serviceName)
		return nil
	}
}
