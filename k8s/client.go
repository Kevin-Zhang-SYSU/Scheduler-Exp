package k8s

/* client.go中定义了对clientSet和discoveryClient以及dynamicClient的创建
这个文件叫client意思是K8S client，clientSet、discoveryClient和dynamicClient都是基于一个Rest Client的*/

import (
	"flag"
	"log"
	"path/filepath"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

// 定义一个K8sConfig类来封装各种方法（这不是一个空类，只是没有成员对象，go语言中的类成员函数是以方法的形式写在类外的）
type K8sConfig struct {
}

func NewK8sConfig() *K8sConfig {
	return &K8sConfig{} //创建一个K8sConfig类型的实例初始化为空值并返回指向该实例的指针
	//然后通过这个指针来调用K8sConfig的方法
}

// 读取kubeconfig 配置文件，这个配置文件其实就是/etc/kubernetes/admin.conf。这里我将其cp到了/$HOME/.kube/config
func (K8sConfig *K8sConfig) K8sRestConfig() *rest.Config {
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err)
	}

	return config //restConfig是创建clientSet,dynamicClient和discoveryClient的基础
}

// 初始化 clientSet
func (K8sConfig *K8sConfig) InitClient() *kubernetes.Clientset {
	clientSet, err := kubernetes.NewForConfig(K8sConfig.K8sRestConfig())

	if err != nil {
		log.Fatal(err)
	}

	return clientSet
}

// 声明全局变量，不能使用:=来定义，要用var。
// 这里有一个潜在的问题就是， K8sRestConfig()这个函数里只能执行一次config的构建
// 多次调用这个函数就会报错flag redefined: kubeconfig
// 所以其实上面这种面向对象的写法感觉不太好，如果后面要扩展出discoveryClient和dynamicClient可能就要修改了。
var ClientSet = NewK8sConfig().InitClient()
