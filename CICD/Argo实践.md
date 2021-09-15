
Argo实践

    简单的一个CI/CD流程实践: 拉代码->编译->构建镜像->上传镜像->部署->自动发布

[TOC]

# 环境准备

## argo-workflows

安装后的资源状态

```shell
[root@k8s-ops argo]# kubectl get all -n argo
NAME                                                 READY   STATUS      RESTARTS   AGE
pod/argo-server-576b68c7cf-jsx8p                     1/1     Running     1          3d1h
pod/minio-58977b4b48-5kx6x                           1/1     Running     0          3d1h
pod/postgres-6b5c55f477-s4njd                        1/1     Running     0          3d1h
pod/workflow-controller-6587f8545-2ct5v              1/1     Running     2          3d1h

NAME                                  TYPE        CLUSTER-IP       EXTERNAL-IP   PORT(S)    AGE
service/argo-server                   ClusterIP   10.254.117.62    <none>        2746/TCP   3d1h
service/minio                         ClusterIP   10.254.34.52     <none>        9000/TCP   3d1h
service/postgres                      ClusterIP   10.254.86.143    <none>        5432/TCP   3d1h
service/workflow-controller-metrics   ClusterIP   10.254.151.196   <none>        9090/TCP   3d1h

NAME                                  READY   UP-TO-DATE   AVAILABLE   AGE
deployment.apps/argo-server           1/1     1            1           3d1h
deployment.apps/minio                 1/1     1            1           3d1h
deployment.apps/postgres              1/1     1            1           3d1h
deployment.apps/workflow-controller   1/1     1            1           3d1h

NAME                                            DESIRED   CURRENT   READY   AGE
replicaset.apps/argo-server-576b68c7cf          1         1         1       3d1h
replicaset.apps/minio-58977b4b48                1         1         1       3d1h
replicaset.apps/postgres-6b5c55f477             1         1         1       3d1h
replicaset.apps/workflow-controller-6587f8545   1         1         1       3d1h
```

## argo-cd

安装后的资源状态

```shell
[root@k8s-ops argo]# kubectl get all -n argocd
NAME                                     READY   STATUS    RESTARTS   AGE
pod/argocd-application-controller-0      1/1     Running   0          7d6h
pod/argocd-dex-server-5bbd5cb9d8-967hk   1/1     Running   0          7d6h
pod/argocd-redis-747b678f89-hmff4        1/1     Running   0          7d6h
pod/argocd-repo-server-7cbbc986b-l5vbl   1/1     Running   0          7d6h
pod/argocd-server-674bc46975-f45nb       1/1     Running   0          7d6h

NAME                            TYPE        CLUSTER-IP       EXTERNAL-IP   PORT(S)                      AGE
service/argocd-dex-server       ClusterIP   10.254.120.132   <none>        5556/TCP,5557/TCP,5558/TCP   7d6h
service/argocd-metrics          ClusterIP   10.254.124.28    <none>        8082/TCP                     7d6h
service/argocd-redis            ClusterIP   10.254.40.173    <none>        6379/TCP                     7d6h
service/argocd-repo-server      ClusterIP   10.254.200.184   <none>        8081/TCP,8084/TCP            7d6h
service/argocd-server           ClusterIP   10.254.63.146    <none>        80/TCP,443/TCP               7d6h
service/argocd-server-metrics   ClusterIP   10.254.156.123   <none>        8083/TCP                     7d6h

NAME                                 READY   UP-TO-DATE   AVAILABLE   AGE
deployment.apps/argocd-dex-server    1/1     1            1           7d6h
deployment.apps/argocd-redis         1/1     1            1           7d6h
deployment.apps/argocd-repo-server   1/1     1            1           7d6h
deployment.apps/argocd-server        1/1     1            1           7d6h

NAME                                           DESIRED   CURRENT   READY   AGE
replicaset.apps/argocd-dex-server-5bbd5cb9d8   1         1         1       7d6h
replicaset.apps/argocd-redis-747b678f89        1         1         1       7d6h
replicaset.apps/argocd-repo-server-7cbbc986b   1         1         1       7d6h
replicaset.apps/argocd-server-674bc46975       1         1         1       7d6h
replicaset.apps/argocd-server-84cb57bf97       0         0         0       7d6h

NAME                                             READY   AGE
statefulset.apps/argocd-application-controller   1/1     7d6h
```

# 开始

## 工作流(argo-workflows)

### 准备

创建好pv, pvc

```shell
[root@k8s-ops argo]# kubectl get pv
NAME             CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS      CLAIM                  STORAGECLASS   REASON   AGE
deploy-work-pv   5Gi        RWX            Retain           Bound       argo/deploy-work-pvc                           24h

[root@k8s-ops argo]# kubectl get pvc -n argo
NAME              STATUS   VOLUME           CAPACITY   ACCESS MODES   STORAGECLASS   AGE
deploy-work-pvc   Bound    deploy-work-pv   5Gi        RWX                           24h
```

定义PVC，用户整个过程的共享存储

```yaml
  # 引用PVC
  volumes:
  - name: deploy-work-dir
    persistentVolumeClaim:
      claimName: deploy-work-pvc
......
    # 使用PVC
    - name: XXX
      script:
        ....
        volumeMounts:
          - mountPath: /work
            name: deploy-work-dir
```

### 拉取代码

使用自定义的镜像`git-setimage:v0.2`拉取代码

```yaml
    - name: Checkout
      script:
        image: k8s-harbor.com/library/git-setimage:v0.2
        workingDir: /work
        command: [sh]
        source: |
          echo "step: Checkout"
          git pull origin {{workflow.parameters.branch}}
        volumeMounts:
          - mountPath: /work
            name: deploy-work-dir
```

构建自定义的镜像`git-setimage:v0.2`

- Dockerfile文件

```shell
FROM alpine:latest
MAINTAINER suqi <git,setimage>

RUN echo "https://mirror.tuna.tsinghua.edu.cn/alpine/v3.13/main" > /etc/apk/repositories
RUN apk add git && rm -rf /var/cache/apk/*
COPY setimage /usr/bin
```

- setimage文件

```shell
[root@k8s-ops git-setimage]# cat setimage 
#!/bin/sh

set -e

if [[ "$1" = "" || "$2" = "" ]];then
  echo "HELP: $0 IMAGE YAMLFILE"
  exit 1
fi

repl="image: "
echo "COMMAND: $0 $1 $2"
grep $repl $2

sed -i "s#$repl.*#image: $1#g" $2 > /dev/null
echo "NEW IMAGE: $1"
```

### 编译打包

使用golang环境，编译打包golang程序

```yaml
    - name: Build
      script:
        image: k8s-harbor.com/library/golang:alpine3.14
        workingDir: /work
        command: [sh]
        source: |
          echo "step: Build"
          go build -o gateway
        volumeMounts:
          - mountPath: /work
            name: deploy-work-dir
```

### 构建镜像

使用[kaniko](https://github.com/GoogleContainerTools/kaniko)来创建镜像

```yaml
    - name: BuildImage
      hostAliases:
        - hostnames: ["k8s-harbor.com"]
          ip: "192.168.101.200"
      volumes:
        - name: docker-config
          secret:
            secretName: docker-config        
      container:
        image: k8s-harbor.com/library/kaniko-project/executor:latest
        workingDir: /work
        args:
          - --context=.
          - --dockerfile=Dockerfile
          - --destination={{workflow.parameters.app-image}}
          - --skip-tls-verify
          - --reproducible
          - --cache=true
        volumeMounts:
          - name: deploy-work-dir
            mountPath: /work
          - name: docker-config
            mountPath: /kaniko/.docker/
```

- `Dockerfile`文件，用于构建镜像

```shell
FROM k8s-harbor.com/library/alpine:latest
COPY gateway /usr/bin
```

生成secret：docker-config(`/root/.docker/config.json`)，并挂载到容器里，用于推送镜像

```shell
kubectl create secret generic docker-config --from-file=/root/.docker/config.json -n argo
```

### 部署

拉取部署的git仓库`argocd-example-apps.git`，修改镜像版本，并提交。提交后的部署yaml文件，用于argo-cd自动发布

```yaml
    - name: Deploy
      script:
        image: k8s-harbor.com/library/git-setimage:v0.2
        command: [sh]
        source: |
          echo "step: Deploy"
          git config --global user.name "devops"
          git config --global user.email "422819555@qq.com"
          git clone --branch master http://carter115:xxxxxx@{{workflow.parameters.deployment-cd-repo}} 
          cd argocd-example-apps/demoapp
          setimage {{workflow.parameters.app-image}} *.yaml
          git add *
          git commit -am "update yaml"
          git push -u origin master
```

### 完整配置

```yaml
apiVersion: argoproj.io/v1alpha1
kind: WorkflowTemplate
metadata:
  annotations:
    workflows.argoproj.io/description: |
      Checkout out from Git, build and deploy application.
    workflows.argoproj.io/maintainer: '@suqi'
    workflows.argoproj.io/tags: golang, git
    workflows.argoproj.io/version: '>= 0.1'
  name: deployment-app 
spec:
  entrypoint: main
  volumes:
  - name: deploy-work-dir
    persistentVolumeClaim:
      claimName: deploy-work-pvc

  arguments:
    parameters:
      - name: repo
        value: https://gitee.com/carter115/app-demo.git
      - name: branch
        value: master
      - name: build-image
        value: k8s-harbor.com/library/golang:alpine3.14
      - name: app-image
        value: k8s-harbor.com/demo/app:0.3
      - name: deployment-cd-repo
        value: gitee.com/carter115/argocd-example-apps.git
      - name: cd-git-username
        value: abc
      - name: cd-git-password
        value: xxxx

  templates:
    - name: main
      inputs:
        parameters:
          - name: appname
      steps:
        - - name: Checkout
            template: Checkout
        - - name: Build
            template: Build
        - - name: BuildImage
            template: BuildImage
        - - name: Deploy
            template: Deploy

    # 1. 拉取代码
    - name: Checkout
      script:
        image: k8s-harbor.com/library/git-setimage:v0.2
        workingDir: /work
        command: [sh]
        source: |
          echo "step: Checkout"
          git pull origin {{workflow.parameters.branch}}
        volumeMounts:
          - mountPath: /work
            name: deploy-work-dir

    # 2. 编译打包
    - name: Build
      script:
        image: k8s-harbor.com/library/golang:alpine3.14
        workingDir: /work
        command: [sh]
        source: |
          echo "step: Build"
          go build -o gateway
        volumeMounts:
          - mountPath: /work
            name: deploy-work-dir

    # 3. 构建镜像
    - name: BuildImage
      hostAliases:
        - hostnames: ["k8s-harbor.com"]
          ip: "192.168.101.200"
      volumes:
        - name: docker-config
          secret:
            secretName: docker-config        
      container:
        image: k8s-harbor.com/library/kaniko-project/executor:latest
        workingDir: /work
        args:
          - --context=.
          - --dockerfile=Dockerfile
          - --destination={{workflow.parameters.app-image}}
          - --skip-tls-verify
          - --reproducible
          - --cache=true
        volumeMounts:
          - name: deploy-work-dir
            mountPath: /work
          - name: docker-config
            mountPath: /kaniko/.docker/

    # 4. 部署
    - name: Deploy
      script:
        image: k8s-harbor.com/library/git-setimage:v0.2
        command: [sh]
        source: |
          echo "step: Deploy"
          git config --global user.name "devops"
          git config --global user.email "422819555@qq.com"
          git clone --branch master http://carter115:xxxx@{{workflow.parameters.deployment-cd-repo}} 
          cd argocd-example-apps/demoapp
          setimage {{workflow.parameters.app-image}} *.yaml
          git add *
          git commit -am "update yaml"
          git push -u origin master


```

## 自动发布(argo-cd)

通过argo-cd的UI创建APP(`demoapp`)

```yaml
source:
  repoURL: 'https://gitee.com/carter115/argocd-example-apps.git'
  path: demoapp
  targetRevision: HEAD
destination:
  server: 'https://kubernetes.default.svc'
  namespace: default
```

发布后的资源

```shell
[root@k8s-ops git-setimage]# kubectl get pod 
NAME                                    READY   STATUS    RESTARTS   AGE
demoapp-6b7685d8f9-5748n                1/1     Running   0          119m
demoapp-6b7685d8f9-5rbbk                1/1     Running   0          119m

[root@k8s-ops git-setimage]# kubectl get svc
NAME                   TYPE        CLUSTER-IP       EXTERNAL-IP   PORT(S)    AGE
demoapp                ClusterIP   10.254.48.180    <none>        8080/TCP   3h11m

[root@k8s-ops git-setimage]# kubectl get IngressRoute
NAME                   AGE
demoapp                3h11m
```

#  参考

  - [argo-workflows](https://argoproj.github.io/argo-workflows/)
  - [argo-cd](https://argoproj.github.io/argo-cd/)
  - [Argo Workflows-Kubernetes的工作流引擎](https://zhuanlan.zhihu.com/p/356240677)
