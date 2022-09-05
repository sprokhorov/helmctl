/*
 Kubernetes

 This module provides functions for Kubernetes interactions

*/
package kubernetes

import (
    "context"
    "fmt"
    "path/filepath"

    "github.com/sirupsen/logrus"
    apiv1 "k8s.io/api/core/v1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/client-go/kubernetes"
    _ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
    "k8s.io/client-go/tools/clientcmd"
    "k8s.io/client-go/util/homedir"
)

// GetKubernetesClient returns Kubernetes client
func GetKubernetesClient(kubeconfigPath string) (*kubernetes.Clientset, error) {
    if kubeconfigPath == "" {
        if home := homedir.HomeDir(); home != "" {
            kubeconfigPath = filepath.Join(home, ".kube", "config")
        }
    }

    config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)

    if err != nil {
        return nil, fmt.Errorf("cannot read kubeconfig file : %v", err)
    }

    clientset, err := kubernetes.NewForConfig(config)
    if err != nil {
        return nil, fmt.Errorf("cannot create kube client : %v", err)
    }

    return clientset, err
}

// checkNamespace checks existence of Namespace and creates it if it's needed
// kubernetes.Interface is used for mock in tests
func CheckNamespace(clientset kubernetes.Interface, name string, dryrun bool) error {
    ctx := context.Background()

    _, err := clientset.CoreV1().Namespaces().Get(ctx, name, metav1.GetOptions{})
    if err == nil {
        logrus.Infof("Namespace %s exists", name)
        return nil
    }
    newNamespace := &apiv1.Namespace{
        ObjectMeta: metav1.ObjectMeta{
            Name: name,
        },
    }

    if !dryrun {
        _, err = clientset.CoreV1().Namespaces().Create(ctx, newNamespace, metav1.CreateOptions{})
        if err != nil {
            return fmt.Errorf("cannot create namespace %s : %v", name, err)
        }
    } else {
        logrus.Infof("Namespace %s will be created", name)
    }

    return nil
}
