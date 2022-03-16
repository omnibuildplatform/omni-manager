package models

import (
	"omni-manager/util"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	clientset *kubernetes.Clientset
)

func Int32Ptr(i int32) *int32 { return &i }
func BoolPtr(i bool) *bool    { return &i }

func InitDispatcherMonitor() {

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", "./docs/infra-test.yaml")
	if err != nil {
		util.Log.Errorln("k8s config error:", err.Error())
		return
	}

	// create the clientset
	clientset, err = kubernetes.NewForConfig(config)
	if err != nil {
		util.Log.Errorln("k8s NewForConfig error:", err.Error())
		return
	}

}
func GetClientSet() *kubernetes.Clientset {
	return clientset
}
