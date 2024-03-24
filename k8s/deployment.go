package k8s

/*对kind:deployment的服务的操作*/

import (
	"context"
	"scheduler-exp/model"
	"scheduler-exp/util"

	"strconv"

	appsv1 "k8s.io/api/apps/v1"
	coreV1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func int32Ptr(i int32) *int32 { return &i }

// 创建一个Deployment
func CreateDeployment(userName string, taskSubmitInfo model.TaskSubmitInfo) error {
	deploymentsClient := ClientSet.AppsV1().Deployments(coreV1.NamespaceDefault)
	image := "registry.cn-hangzhou.aliyuncs.com/k8s-image-linchx/k8s-image:pytorch2.1.0-cuda11.8-cudnn8-runtime-v3"
	//下面创建一个deployment,类似于写一个deploy的yaml文件
	//官方文档见：https://pkg.go.dev/k8s.io/api/apps/v1#Deployment
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: userName + "-" + taskSubmitInfo.TaskName,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"task": userName + "-" + taskSubmitInfo.TaskName,
				},
			},
			Template: coreV1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"task": userName + "-" + taskSubmitInfo.TaskName,
					},
				},
				Spec: coreV1.PodSpec{
					Volumes: []coreV1.Volume{
						{
							Name: "nfs-pvc-vol",
							VolumeSource: coreV1.VolumeSource{
								PersistentVolumeClaim: &coreV1.PersistentVolumeClaimVolumeSource{
									ClaimName: "nfs-static-pvc",
								},
							},
						},
					},
					Containers: []coreV1.Container{ //注意这里是数组类型，一个deployment是可以创建多个不同容器的
						{ //这层大括号是创建一个匿名结构体
							Name:  "jupyterlab-codeserver-container",
							Image: image,
							Ports: []coreV1.ContainerPort{
								{
									Name:          "jupyterlab",
									Protocol:      coreV1.ProtocolTCP,
									ContainerPort: 8888,
								},
								{
									Name:          "codeserver",
									Protocol:      coreV1.ProtocolTCP,
									ContainerPort: 8889,
								},
							},
							VolumeMounts: []coreV1.VolumeMount{
								{
									Name:      "nfs-pvc-vol",
									MountPath: "/root/jupyter_workspace/dataset",
								},
							},
							Env: []coreV1.EnvVar{
								{
									Name:  "USER_PASSWORD",
									Value: taskSubmitInfo.Password,
								},
							},
							Resources: coreV1.ResourceRequirements{
								Limits: coreV1.ResourceList{
									"aliyun.com/gpu-mem": resource.MustParse(strconv.Itoa(taskSubmitInfo.Gpu)),
								},
							},
						},
					},
					ImagePullSecrets: []coreV1.LocalObjectReference{
						{
							Name: "registry-secret", //deployment要和secret在同一命名空间
						},
					},
				},
			},
		},
	}
	util.Logger.Debug("Creating deployment ... ... ...")
	result, err := deploymentsClient.Create(context.TODO(), deployment, metav1.CreateOptions{})
	if err != nil {
		util.Logger.Errorf("Failed to create deployment %q: %v", deployment.GetObjectMeta().GetName(), err)
		return err
	}
	util.Logger.Debugf("Created deployment %q.", result.GetObjectMeta().GetName())
	return nil
}

// 获取一个Deployment
func GetDeployment(deploymentName string) *appsv1.Deployment {
	namespace := "default"
	deployment, err := ClientSet.AppsV1().Deployments(namespace).Get(context.TODO(), deploymentName, metav1.GetOptions{})
	if err != nil {
		util.Logger.Errorf("Failed to get deployment %q: %v", deploymentName, err)
		return nil
	} else {
		util.Logger.Debugf("Get deployment %q.", deploymentName)
		return deployment
	}
}

// 删除一个Deployment
func DeleteDeployment(deploymentName string) error {
	namespace := "default"
	err := ClientSet.AppsV1().Deployments(namespace).Delete(context.TODO(), deploymentName, metav1.DeleteOptions{})
	if err != nil {
		util.Logger.Errorf("Failed to delete deployment %q: %v", deploymentName, err)
		return err
	} else {
		util.Logger.Debugf("Deleted deployment %q.", deploymentName)
		return nil
	}
}
