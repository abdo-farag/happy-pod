package main

import (
	"context"
	"fmt"
	"time"
	"strconv"
	"strings"

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
				if err := protect(clientset, pod); err != nil {
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

func protect(clientset kubernetes.Interface, pod *corev1.Pod) error {
	egressPorts := []struct {
        port     int
        protocol v1.Protocol
    }{}
    if val, ok := pod.Annotations["egress-ports-tcp"]; ok {
        portStrings := strings.Split(val, ",")
        for _, portString := range portStrings {
            port, err := strconv.Atoi(portString)
            if err != nil {
                return err
            }
            egressPorts = append(egressPorts, struct {
                port     int
                protocol v1.Protocol
            }{
                port:     port,
                protocol: v1.ProtocolTCP,
            })
        }
    }
    if val, ok := pod.Annotations["egress-ports-udp"]; ok {
        portStrings := strings.Split(val, ",")
        for _, portString := range portStrings {
            port, err := strconv.Atoi(portString)
            if err != nil {
                return err
            }
            egressPorts = append(egressPorts, struct {
                port     int
                protocol v1.Protocol
            }{
                port:     port,
                protocol: v1.ProtocolUDP,
            })
        }
    }

	networkPolicy := &networkingv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "protect-"+pod.Name,
			Namespace: pod.Namespace,
		},
		Spec: networkingv1.NetworkPolicySpec{
			PodSelector: metav1.LabelSelector{
				MatchLabels: map[string]string{
					"run": pod.Labels["run"],
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
					Ports: []networkingv1.NetworkPolicyPort{},
				},
			},
		},
	}

	for _, port := range egressPorts {
        networkPolicy.Spec.Egress[0].Ports = append(networkPolicy.Spec.Egress[0].Ports, networkingv1.NetworkPolicyPort{
            Port:     &[]intstr.IntOrString{intstr.FromInt(port.port)}[0],
            Protocol: &port.protocol,
        })
    }
	if _, err := clientset.NetworkingV1().NetworkPolicies("default").Create(
		context.Background(), networkPolicy, metav1.CreateOptions{},
	); err != nil {
		return err
	}

	return nil
}
