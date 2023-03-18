package main

import (
	"context"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/informers"
	coreinformers "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

type PodInformer struct {
	stopCh          chan struct{}
	podInformer     coreinformers.PodInformer
	informerFactory informers.SharedInformerFactory
}

func NewPodInformer(clientset kubernetes.Interface) (*PodInformer, error) {
	informerFactory := informers.NewSharedInformerFactory(clientset, time.Hour*24)
	podInformer := informerFactory.Core().V1().Pods()

	podInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			pod := obj.(*corev1.Pod)
			if pod.Annotations[foo] == bar {
				if err := protect(clientset); err != nil {
					fmt.Printf("fail to protect pod %v", pod.Name)
				}
			}
		},
	})

	return &PodInformer{
		podInformer:     podInformer,
		informerFactory: informerFactory,
	}, nil
}

func (p *PodInformer) Start() {
	p.stopCh = make(chan struct{})
	p.informerFactory.Start(p.stopCh)
	<-p.stopCh
}

func (p *PodInformer) Stop() {
	defer runtime.HandleCrash()
	close(p.stopCh)
}

func protect(clientset kubernetes.Interface) error {
	networkPolicy := &networkingv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "protect-my-pod",
			Namespace: "default",
		},
		Spec: networkingv1.NetworkPolicySpec{
			PodSelector: metav1.LabelSelector{
				MatchLabels: map[string]string{
					"run": "foo",
				},
			},
			PolicyTypes: []networkingv1.PolicyType{
				networkingv1.PolicyTypeEgress,
			},
			Egress: []networkingv1.NetworkPolicyEgressRule{
				{
					To: []networkingv1.NetworkPolicyPeer{
						{
							IPBlock: &networkingv1.IPBlock{
								CIDR: "0.0.0.0/0",
							},
						},
					},
					Ports: []networkingv1.NetworkPolicyPort{
						{
							Port:     &[]intstr.IntOrString{intstr.FromInt(80)}[0],
							Protocol: &[]v1.Protocol{v1.ProtocolTCP}[0],
						},
						{
							Port:     &[]intstr.IntOrString{intstr.FromInt(443)}[0],
							Protocol: &[]v1.Protocol{v1.ProtocolTCP}[0],
						},
					},
				},
			},
		},
	}

	if _, err := clientset.NetworkingV1().NetworkPolicies("default").Create(
		context.Background(), networkPolicy, metav1.CreateOptions{},
	); err != nil {
		return err
	}

	return nil
}
