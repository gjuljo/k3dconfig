# RUNNING K3D
Thanks to [k3d](https://github.com/rancher/k3d), you can run a Kubernetes cluster in your laptop by just using docker.

## DEMO APPLICATION
In this tutorial we use a simple Go application that just listes to incomping HTTP requests on port `8001`, returing the content of the file `static/hello.html`.
This file can three templates variables:
- `{{.Start}}`
- `{{.Username}}`
- `{{.End}}`

The actual value of those template variales can be set with the corresponding environment variables:
- `MYAPP_START`, default value `mystart`
- `MYAPP_USERNAME`, default value `myuser`
- `MYAPP_END`, default value `myend`

You can combine those variables at your will to generate the message you want, by just using them in `hello.html`, whose default content is:

```
default --> {{.Start}} {{.Username}} {{.End}}
```

So, by default, this is the message generated by the application:

```
default --> mystart myuser myend
```

## GENERATE THE DOCKER IMAGE
First of all build the docker image: open a prompt in the `goapp` directory and run the following command:

```console
$> docker build -t gjuljo/myapp .
```

Of course you are free to change the image name, but please remember to update all the subsequent commands below.

### TEST THE IMAGE WITH DEFAULT VALUES
Go back to the root directory of the project and test the docker image:

```console
$> docker run -it --rm -p 8001:8001 gjuljo/myapp
```

```console
$> curl localhost:8001
default --> mystart myuser myend
```

### TEST THE IMAGE WITH CUSTOM ENVIRONMENT VARIABLES
Stop the running image and rerun it by providing values for the environment variables:

```console
$> docker run -it --rm -p 8001:8001 -e MYAPP_START=Hello -e MYAPP_USERNAME=Giulio -e MYAPP_END=", how are you?" gjuljo/myapp
```

```console
$> curl localhost:8001
default --> Hello Giulio how are you?
```

### TEST THE IMAGE WITH CUSTOM ENVIRONMENT VARIABLES AND VOLUME
Stop the running image and, in addition to the custom environment variables, replace the hello.html file with an external volumen::

On Windows:
```console
$> docker run -it --rm -p 8001:8001 -v %CD%/hello1/:/app/static/ -e MYAPP_START=Hello -e MYAPP_USERNAME=Giulio -e MYAPP_END="how are you?" gjuljo/myapp
```

On Linux
```console
$> docker run -it --rm -p 8001:8001 -v $pwd/hello1/:/app/static/ -e MYAPP_START=Hello -e MYAPP_USERNAME=Giulio -e MYAPP_END="how are you?" gjuljo/myapp
```

```console
$> curl localhost:8001
hello1 --> Hello Giulio how are you?
```


## CREATE K3D CLUSTER WITH A LOCAL CONTAINER REGISTRY
Let's create a local registry (i.e. the container `registry.local`) that you access from your host using a local hostname (`registry.lvh.me`). 

1. Create a volume to host the registry:

```console
$> docker volume create local_registry
```

2. Create a container running the registry image:

```console
$> docker container run -d --name registry.local -v local_registry:/var/lib/registry --restart always -p 5000:5000 registry:2
```

3. Tag your image and publish it to the local registry. **You should do it every time you change the image contents**:

```console
$> docker tag gjuljo/myapp:latest registry.lvh.me:5000/gjuljo/myapp:latest
$> docker push registry.lvh.me:5000/gjuljo/myapp:latest
```

4. Create the `registries.yaml` file in your filesystem (ATTENTION: if you are running WSL1, create this file in the Windows filesystem, i.e. `C:\Work\k3d\registry`):

```
mirrors:
  "registry.lvh.me:5000":
    endpoint:
    - http://registry.local:5000
```

Please notice that k3d maps the registry label (`registry.lvh.me:5000`) to the hostname of the container running as registry (`registry.local:5000`).


5. Create the k3d cluster mapping the local `registry` volume, where `registries.yaml` is supposed to be:

```console
$> k3d create -publish 80:80 --volume $PWD/registry:/etc/rancher/k3s
```

6. Export the Kubernetes configuration file and wait for the cluster to be up and running

```console
$> export KUBECONFIG=$(k3d get-kubeconfig)
```

7. Connect the local registry (i.e. the container `registry.local`) to the docker netword created by k3d:

```console
$> docker network connect k3d-k3s-default registry.local
```

This must be done at least once, as this connection can be reused every time you delete and create a k3d cluster on the same environment, unless you delete it.


## TEST KUBERNETES DEPLOYMENT FILES
In the first Kubernetes example, we just create a `Deployemnt`, a `Service` and an `Ingress` object:

```
apiVersion: apps/v1
kind: Deployment
metadata:
  name: test1-deployment
  labels:
    app: test1-app
spec:
  replicas: 1
  selector:
    matchLabels:
      app: test1-app
  template:
    metadata:
      labels:
        app: test1-app
    spec:
      containers:
      - name: test1-app
        image: registry.lvh.me:5000/gjuljo/myapp:latest
        ports:
        - containerPort: 8001
        env:
        - name: MYAPP_USERNAME
          value: Giulio
---
apiVersion: v1
kind: Service
metadata:
  name: test1-service
  labels:
    app: test1-service
spec:
  ports:
  - port: 8001
    targetPort: 8001
    name: http
  selector:
    app: test1-app
---
apiVersion: extensions/v1beta1
kind: Ingress
metadata:
  name: test1-ingress
  annotations:
    ingress.kubernetes.io/ssl-redirect: "false"
spec:
  rules:
  - host: test1.lvh.me
    http:
      paths:
      - backend:
          serviceName: test1-service
          servicePort: 8001
```

Notice the following:
- the image name referes to the local registry;
- the environment variable `MYAPP_USERNAME` is set directly in the `Deployment` object without using any additional indirection mechanism (i.e. a `ConfigMap`);
- the `Ingress` object exports the hostname `test1.lvh.me` (automatically resolved as `127.0.0,1`) makes the local ingress to listen at the Kubernetes port (i.e. `80`).

```console
$> kubectl create -f test1-default.yaml
```

This is what you get when you invoke the service:

```console
$> curl test1.lvh.me
default --> mystart Giulio myend
```

You get the same result even if you use the ip address (`127.0.0.1`) and set the `Host` header with the expected hostname (i.e. `test1.lvh.me`):

```console
$> curl -H 'Host:test1.lvh.me' 127.0.0.1 
default --> mystart Giulio myend
```

## TEST A CONFIG MAP FOR THE ENVIRONMENT VARIABLES
Now we also add a `ConfigMap` to set the value for the other environment variables, `MYAPP_START` and `MYAPP_END`, that is referred in the `Deploymenet` object:

```
kind: ConfigMap
apiVersion: v1
metadata:
  name: test2-config
data:
  MYAPP_START_CONFIG: "Hello"
  MYAPP_END_CONFIG: "how are you?"
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: test2-deployment
  labels:
    app: test2-app
spec:
  replicas: 1
  selector:
    matchLabels:
      app: test2-app
  template:
    metadata:
      labels:
        app: test2-app
    spec:
      containers:
      - name: test2-app
        image: registry.lvh.me:5000/gjuljo/myapp:latest
        ports:
        - containerPort: 8001
        env:
        - name: MYAPP_USERNAME
          value: Giulio
        - name: MYAPP_START
          valueFrom:
            configMapKeyRef:
              name: test2-config
              key: MYAPP_START_CONFIG
        - name: MYAPP_END
          valueFrom:
            configMapKeyRef:
              name: test2-config
              key: MYAPP_END_CONFIG     
---
apiVersion: v1
kind: Service
metadata:
  name: test2-service
  labels:
    app: test2-service
spec:
  ports:
  - port: 8001
    targetPort: 8001
    name: http
  selector:
    app: test2-app
---
apiVersion: extensions/v1beta1
kind: Ingress
metadata:
  name: test2-ingress
  annotations:
    ingress.kubernetes.io/ssl-redirect: "false"
spec:
  rules:
  - host: test2.lvh.me
    http:
      paths:
      - backend:
          serviceName: test2-service
          servicePort: 8001
```

In this second deployment, the environment variables `MYAPP_START` and `MYAPP_END` are, correspondigly, `Hello` and `how are you?`:

```console
$> kubectl create -f test2-env.yaml
```

In this second test the ingress hostname is `test2.lvh.me`:

```console
$> curl test2.lvh.me
default --> Hello Giulio how are you?
```


## TEST A CONFIG MAP FOR THE HTML FILE
You can even provide the contents of the `hello.html` file, by using the same or additional `ConfigMap` that contans the content of the file itself and that can be mounted as a `Volume`:

```console
kind: ConfigMap
apiVersion: v1
metadata:
  name: test3-config-vol
data:
  hello.html: |
     hello --> {{.Start}} {{.Username}} {{.End}}
---
kind: ConfigMap
apiVersion: v1
metadata:
  name: test3-config-env
data:
  MYAPP_START: "Hello"
  MYAPP_END: "how are you?"
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: test3-deployment
  labels:
    app: test3-app
spec:
  replicas: 1
  selector:
    matchLabels:
      app: test3-app
  template:
    metadata:
      labels:
        app: test3-app
    spec:
      containers:
      - name: test3-app
        image: registry.lvh.me:5000/gjuljo/myapp:latest
        ports:
        - containerPort: 8001
        env:
        - name: MYAPP_USERNAME
          value: Giulio
        - name: MYAPP_START
          valueFrom:
            configMapKeyRef:
              name: test3-config-env
              key: MYAPP_START
        - name: MYAPP_END
          valueFrom:
            configMapKeyRef:
              name: test3-config-env
              key: MYAPP_END
        volumeMounts:
        - name: test3-vol
          mountPath: /app/static                  
      volumes:
      - name: test3-vol
        configMap:
          name: test3-config-vol
---
apiVersion: v1
kind: Service
metadata:
  name: test3-service
  labels:
    app: test3-service
spec:
  ports:
  - port: 8001
    targetPort: 8001
    name: http
  selector:
    app: test3-app
---
apiVersion: extensions/v1beta1
kind: Ingress
metadata:
  name: test3-ingress
  annotations:
    ingress.kubernetes.io/ssl-redirect: "false"
spec:
  rules:
  - host: test3.lvh.me
    http:
      paths:
      - backend:
          serviceName: test3-service
          servicePort: 8001
```

Once the service is created:

```console
> kubectl create -f test3-vol.yaml
```

the usual invocation generates, this time, a different content:

```console
> curl test3.lvh.me
hello: Hello Giulio how are you?
```


# REFERENCES
[k3d](https://github.com/rancher/k3d)
[k3sup](https://github.com/alexellis/k3sup)
[k3s](https://k3s.io/)
