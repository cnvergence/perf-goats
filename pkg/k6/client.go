package k6

import (
	"context"
	"errors"
	"log"
	"os"
	"path/filepath"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

type Config struct {
	ClientSet     kubernetes.Interface
	RestConfig    *rest.Config
	Context       context.Context
	PodName       string
	ContainerName string
	Namespace     string
}

func NewConfig(ctx context.Context) *Config {
	clientSet, restConfig := initK8sClient()
	return &Config{ClientSet: clientSet, RestConfig: restConfig, Context: ctx, PodName: "goat-test-k6-d575d4d67-sxfg6", ContainerName: "k6-alpine", Namespace: "load-test"}
}

func initK8sClient() (kubernetes.Interface, *rest.Config) {
	var kubeconfig string
	if kConfig, ok := os.LookupEnv("KUBECONFIG"); !ok {
		if home := homedir.HomeDir(); home != "" {
			kubeconfig = filepath.Join(home, ".kube", "config")
		}
	} else {
		kubeconfig = kConfig
	}
	_, err := os.Stat(kubeconfig)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			log.Fatalf("kubeconfig %s does not exist", kubeconfig)
		}
		log.Fatalf(err.Error())
	}
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		log.Fatalf(err.Error())
	}
	k8sClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf(err.Error())
	}

	return k8sClient, config
}
