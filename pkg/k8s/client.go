package k8s

import (
	"context"
	"fmt"

	"github.com/cyberark/conjur-authn-k8s-client/pkg/log"
	"github.com/cyberark/conjur-k8s-csi-provider/pkg/logmessages"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type GetPodAnnotationsFunc func(namespace string, podName string) (map[string]string, error)

func GetPodAnnotations(namespace string, podName string) (map[string]string, error) {
	kubeClient, _ := configK8sClient()

	pod, err := kubeClient.CoreV1().Pods(namespace).Get(context.Background(), podName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf(logmessages.CKCP039, podName, namespace, err.Error())
	}

	return pod.Annotations, nil
}

func configK8sClient() (*kubernetes.Clientset, error) {
	log.Info(logmessages.CKCP036)
	kubeConfig, err := rest.InClusterConfig()
	if err != nil {
		log.Error(logmessages.CKCP037, err.Error())
		return nil, fmt.Errorf(logmessages.CKCP037, err.Error())
	}

	kubeClient, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		log.Error(logmessages.CKCP038, err.Error())
		return nil, fmt.Errorf(logmessages.CKCP038, err.Error())
	}

	return kubeClient, nil
}
