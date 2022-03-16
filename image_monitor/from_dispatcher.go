package image_monitor

import (
	"context"
	"fmt"
	"omni-manager/util"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	clientset *kubernetes.Clientset
)

func StartDispatcherMonitor() {

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
	for {

		pods, err := clientset.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			util.Log.Errorln("k8s Pods List error :", err.Error())
			return
		}
		fmt.Printf("There are %d pods in the cluster\n", len(pods.Items))

		namespace := "default"
		pod := "image-builder"
		_, err = clientset.CoreV1().Pods(namespace).Get(context.TODO(), pod, metav1.GetOptions{})
		if errors.IsNotFound(err) {
			fmt.Printf("Pod %s in namespace %s not found\n", pod, namespace)
		} else if statusError, isStatus := err.(*errors.StatusError); isStatus {
			fmt.Printf("Error getting pod %s in namespace %s: %v\n",
				pod, namespace, statusError.ErrStatus.Message)
		} else if err != nil {
			panic(err.Error())
		} else {
			fmt.Printf("Found pod %s in namespace %s\n", pod, namespace)
		}

		time.Sleep(10 * time.Second)
	}
}
func GetClientSet() *kubernetes.Clientset {
	return clientset
}
