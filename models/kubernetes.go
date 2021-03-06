package models

import (

	// _ "github.com/kubernetes/component-helpers"

	"github.com/omnibuildplatform/omni-manager/util"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	clientset *kubernetes.Clientset
	k8sconfig *rest.Config
)

func InitK8sClient() (err error) {
	// use the current context in kubeconfig
	k8sconfig, err = clientcmd.BuildConfigFromFlags("", "./docs/infra-test.yaml")
	if err != nil {
		util.Log.Error("k8s config error:", err.Error())
		return
	}

	// create the clientset
	clientset, err = kubernetes.NewForConfig(k8sconfig)
	if err != nil {
		util.Log.Error("k8s NewForConfig error:", err.Error())
		return
	}
	return nil
}
func GetClientSet() *kubernetes.Clientset {
	return clientset
}

func GetK8sConfig() *rest.Config {
	return k8sconfig
}
