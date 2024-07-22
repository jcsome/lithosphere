package k8sClient

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func NewK8sClient() (client *kubernetes.Clientset, err error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	client, err = kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	return client, nil
}
