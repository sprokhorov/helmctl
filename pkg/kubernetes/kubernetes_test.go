package kubernetes

import (
    "context"
    "testing"
    "time"

    "k8s.io/api/core/v1"
    "k8s.io/apimachinery/pkg/util/wait"
    "k8s.io/client-go/informers"
    "k8s.io/client-go/kubernetes/fake"
    "k8s.io/client-go/tools/cache"
)

// TestCheckNamespace tests creating namespace with mock client
func TestCheckNamespace(t *testing.T) {
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    // Create the fake client.
    client := fake.NewSimpleClientset()

    // We will create an informer that writes added namespaces to a channel.
    namespaces := make(chan *v1.Namespace, 1)
    informers := informers.NewSharedInformerFactory(client, 0)
    namespaceInformer := informers.Core().V1().Namespaces().Informer()
    namespaceInformer.AddEventHandler(&cache.ResourceEventHandlerFuncs{
        AddFunc: func(obj interface{}) {
            namespace := obj.(*v1.Namespace)
            t.Logf("namespace added: %s", namespace.Name)
            namespaces <- namespace
        },
    })

    // Make sure informers are running.
    informers.Start(ctx.Done())

    // This is not required in tests, but it serves as a proof-of-concept by
    // ensuring that the informer goroutine have warmed up and called List before
    // we send any events to it.
    cache.WaitForCacheSync(ctx.Done(), namespaceInformer.HasSynced)

    err := CheckNamespace(client, "fake-namespace", false)
    if err != nil {
        t.Fatalf("error injecting namespace add: %v", err)
    }

    select {
    case namespace := <-namespaces:
        t.Logf("Got namespace from channel: %s", namespace.Name)
    case <-time.After(wait.ForeverTestTimeout):
        t.Error("Informer did not get the added namespace")
    }
}
