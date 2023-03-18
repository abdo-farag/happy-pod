# emnify happy pod

We want to make sure that _FOO_ pod is happy.

## Setup

Follow the steps to have the setup running on your local machine:

1. [install k3d](https://k3d.io/v5.4.7/#installation)
2. create a cluster
```shell
k3d cluster create emnify-test
```
3. install cert-manager
```shell
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.11.0/cert-manager.yaml
```
4. deploy our happy controller
```shell
kubectl create -f deploy.yaml
```

## Test
Let's create FOO pod
```shell
kubectl create -f foo-pod.yaml
```

Is FOO happy? If not, fix it. It must be happy!

You should see FOO pod running and continuously outputting:
```shell
$ k logs -f foo
I am happy!
I am happy!
...
```

What we want with this test:
- Describe why FOO was not happy
- Show us the fix
- Describe the steps used to debug the problem
- Describe the concepts involved in this project
- (Bonus) Improve controller code

------------------------------------------------------------------------
------------------------------------------------------------------------
## Solution
------------------------------------------------------------------------
------------------------------------------------------------------------
- To improve the build speed of my Dockerfile, I decided to minimize it. After doing so, I built the image and pushed it to my Docker registry.

- I then customized the controller deployment to use the image that I had just built. Once the deployment was customized, I deployed the foo-pod.

- When I checked the logs of the foo-pod, I discovered that the pod failed to pull the image "foo/bar".

#### first issue
- I investigated further and found that the reason why the foo-pod failed to start was that the code condition for starting the container did not match. To fix this, I added an annotation key "foo" with the value "I am happy" to the metadata in the foo-pod manifest file. This allowed the controller admission to start the pod using the configured image, rather than replacing it with the nginx image.
```
---
apiVersion: v1
kind: Pod
metadata:
  annotations:
    foo: "I am happy"
```
- After fixing the first issue, I checked the foo-pod logs again and found that the pod failed with the error "curl: (6) Could not resolve host: example.com".

#### second issue
- Upon further investigation, I discovered that the script inside the container was unable to access the example.com host because the Kubernetes network policy egress rule only allowed ports 80 and 443, but not port 53 for domain name resolution functionality.

- To fix this issue, I added an egress rule for port 53 UDP to the network policy. After doing so, I rebuilt the image and redeployed the pod.
```
//...
Ports: []networkingv1.NetworkPolicyPort{
	...
	{
		Port:     &[]intstr.IntOrString{intstr.FromInt(53)}[0],
		Protocol: &[]v1.Protocol{v1.ProtocolUDP}[0],
	},
},
//...
```
#### Problem solved
- Finally, I checked the pod logs again and found that the pod was happy and kept printing "I am happy".

#### the concepts involved in this project
This project involves creating a Kubernetes controler and a webhook server that protects pods with specific annotations.

The informer.go file sets up a shared informer for Kubernetes pods and creates a network policy to restrict egress traffic on pods that have a specific annotation. 

The main.go file creates a Kubernetes controller mannager and sets up a webhook server that will invoke the handler function for every request it receives on the /happy-pod endpoint. The handler function receives a request to mutate a pod and applies a json patch that adds the annnotation foo: "I am happy" to the pod's metadata.

#### Changes made to the controller code

I made some changes to the original code to improve its functionality. First, I modified the protect funnction to create a networkingv1.NetworkPolicy object that specifies a pod selector based on a label selector. This allows the controller to create network policies for all pods that have the same label, instead of just a single pod.

Additionally, I added suppport for different egress ports by adding an array of networkingv1.NetworkPolicyPort objects to the networkiingv1.NetworkPolicyEgressRule object. This allows for greater flexibility in defining network policies that can match different ports.

Then, I added support for creating network policies for both TCP and UDP protocolls. This was done by adding a conditional statement that checks the protocol specified in the annotation against the desired protocol, and only creating a network policy if the protocol matches.

These changes have made the controller more felexible and useful for creating network policies for a wider range of pods with varying requiremments.

#### Another enhancement that could be made
-----------------------------------------------------------------------------------
Useing a ConfigMap to define the ports for labels instead of annotations. This would invovlve creating a ConfigMap that stores the labels as keys and their corresponding ready ports as values.

With this approach, the protect function can be modified to read the ConfigMap, retrieve the ready ports for a label, and dynamically create the egress rules accordingly. This way, it would be possible to change the ready ports for a label without modifying the code or annotatoins.

Using a ConfigMap in this way provides a more efficient strategy for managing network policies across multiple pods with different labels, as it enables the required ports ports to be centrally managed and easily updated without the need to modify individual annotations or code. If time permits, implementing this improvement would be a worthwhile investment.
