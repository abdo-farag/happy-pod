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
