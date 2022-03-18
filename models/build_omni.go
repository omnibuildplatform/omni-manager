package models

import (
	"bytes"
	"context"
	"io"
	"omni-manager/util"

	// _ "github.com/kubernetes/component-helpers"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	clientset *kubernetes.Clientset
	k8sconfig *rest.Config
)

func InitDispatcherMonitor() (err error) {
	// use the current context in kubeconfig
	k8sconfig, err = clientcmd.BuildConfigFromFlags("", "./docs/infra-test.yaml")
	if err != nil {
		util.Log.Errorln("k8s config error:", err.Error())
		return
	}

	// create the clientset
	clientset, err = kubernetes.NewForConfig(k8sconfig)
	if err != nil {
		util.Log.Errorln("k8s NewForConfig error:", err.Error())
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

func getPodLogs(pod corev1.Pod) string {
	podLogOpts := corev1.PodLogOptions{}
	config, err := rest.InClusterConfig()
	if err != nil {
		return "error in getting config"
	}
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return "error in getting access to K8S"
	}
	req := clientset.CoreV1().Pods(pod.Namespace).GetLogs(pod.Name, &podLogOpts)
	podLogs, err := req.Stream(context.TODO())
	if err != nil {
		return "error in opening stream"
	}
	defer podLogs.Close()

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, podLogs)
	if err != nil {
		return "error in copy information from podLogs to buf"
	}
	str := buf.String()

	return str
}
