package main

import (
	"context"
	"encoding/json"
	"log"

	"gomodules.xyz/jsonpatch/v2"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	runtimeConfig "sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

const (
	foo = "foo"
	bar = "I am happy"
)

func main() {
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Fatal("fail to create cluster config: ", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal("failt to create cluster client: ", err)
	}

	podInformer, err := NewPodInformer(clientset)
	if err != nil {
		log.Fatal("unable start pod informer: ", err)
	}

	defer podInformer.Stop()
	go podInformer.Start()

	mgr, err := manager.New(runtimeConfig.GetConfigOrDie(), manager.Options{})
	if err != nil {
		log.Fatal("unable to setup controller manager: ", err)
	}

	log.Print("setting up webhook server")
	hookServer := mgr.GetWebhookServer()

	log.Print("registering webhook endpoint")
	hookServer.Register("/happy-pod", &admission.Webhook{
		Handler: admission.HandlerFunc(handler),
	})

	log.Print("starting manager")
	if err := mgr.Start(signals.SetupSignalHandler()); err != nil {
		log.Fatal("unable to run manager: ", err)
	}
}

func handler(ctx context.Context, req admission.Request) admission.Response {
	log.Printf("admission request pod name %v", req.Name)

	pod := &v1.Pod{}
	if err := json.Unmarshal(req.Object.Raw, pod); err != nil {
		log.Fatal("could not unmarshall request: ", err)
	}

	val, ok := pod.Annotations[foo]
	if ok && val == bar {
		return admission.Allowed("")
	} else {
		return admission.Patched(
			req.Name,
			jsonpatch.Operation{
				Operation: "add",
				Path:      "/spec/containers/0/image",
				Value:     "foo/bar",
			},
		)
	}
}
