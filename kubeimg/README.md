# kubeimg

简单实现一个image命令，用于查看不同资源对象(deployments/daemonsets/statefulsets/jobs/cronjobs)的名称，和对应容器名称，镜像名称．

## 开始

构建

```shell
go build -o kubeimg
```

使用帮助

```shell
./kubeimg --help
Usage:
  kubeimg [command]

Available Commands:
  completion  generate the autocompletion script for the specified shell
  help        Help about any command
  image       show image

Flags:
  -h, --help                help for kubeimg
  -c, --kubeconfig string   kubeconfig file
  -n, --namespace string    k8s namespace (default "default")

Use "kubeimg [command] --help" for more information about a command.
```

作为kubectl插件使用

```shell
# 修改命令
cp kubeimg /usr/local/bin/kubectl-img

# 使用插件
kubectl img image -d -e -n kube-system
+-------------+------------+----------------+----------------+-------------------------------+
|  NAMESPACE  |    TYPE    |      NAME      |     CNAME      |             IMAGE             |
+-------------+------------+----------------+----------------+-------------------------------+
| kube-system | deployment | metrics-server | metrics-server | rancher/metrics-server:v0.3.6 |
| kube-system | deployment | metrics-server | metrics-server | rancher/metrics-server:v0.3.6 |
| kube-system | deployment | metrics-server | metrics-server | rancher/metrics-server:v0.3.6 |
| kube-system | deployment | metrics-server | metrics-server | rancher/metrics-server:v0.3.6 |
| kube-system | daemonset  | svclb-traefik  |  lb-port-443   |   rancher/klipper-lb:v0.2.0   |
| kube-system | daemonset  | svclb-traefik  |  lb-port-443   |   rancher/klipper-lb:v0.2.0   |
+-------------+------------+----------------+----------------+-------------------------------+
```
