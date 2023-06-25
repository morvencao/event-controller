## Event Controller Example

An example for how to implement a controller using cloud event from MQTT.

## Get Started

1. Set up KinD cluster

```shell
cat <<EOF | kind create cluster --config -
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
  extraPortMappings:
  - containerPort: 31320
    hostPort: 31320
    listenAddress: "127.0.0.1"
    protocol: TCP
EOF
```

2. Deploy MQTT server

```shell
kubectl apply -f deploy/mqtt-server.yaml
```

3. Run example controller:

```shell
go run main.go
```

4. Publish message from another terminal:

```shell
mosquitto_pub -h 127.0.0.1 -p 31320 -u admin -P password -t "v1/client01/testing001/content" -m '{ "sentTimestamp": 1680964785, "resourceGenerationID": "12345", "content": { "apiVersion": "apps/v1", "kind": "Deployment", "metadata": { "uid": "123", "name": "nginx", "namespace": "default", "generation": 12 }, "spec": { "replicas": 1, "selector": { "matchLabels": { "app": "nginx" } }, "template": { "metadata": { "labels": { "app": "nginx" } }, "spec": { "containers": [ { "image": "nginx:1.14.2", "name": "nginx" } ] } } } } }' 
```

5. Check the logs from example controller:

```shell
# go run example/main.go
2023-06-25T08:11:13Z	INFO	controller-runtime.metrics	Metrics server is starting to listen	{"addr": ":8080"}
2023-06-25T08:11:13Z	INFO	setup	starting manager
2023-06-25T08:11:13Z	INFO	starting server	{"path": "/metrics", "kind": "metrics", "addr": "[::]:8080"}
2023-06-25T08:11:13Z	INFO	Starting EventSource	{"controller": "sync-controller", "source": "channel source: 0xc0000c9640"}
2023-06-25T08:11:13Z	INFO	Starting Controller	{"controller": "sync-controller"}
2023-06-25T08:11:13Z	INFO	Starting workers	{"controller": "sync-controller", "worker count": 1}
2023-06-25T08:11:13Z	INFO	mqtt-connection	mqtt connection up
2023-06-25T08:11:13Z	INFO	mqtt-connection	subscribed



2023-06-25T08:11:18Z	INFO	mqtt-connection	received: v1/client01/testing001/content { "sentTimestamp": 1680964785, "resourceGenerationID": "12345", "content": { "apiVersion": "apps/v1", "kind": "Deployment", "metadata": { "uid": "123", "name": "nginx", "namespace": "default", "generation": 12 }, "spec": { "replicas": 1, "selector": { "matchLabels": { "app": "nginx" } }, "template": { "metadata": { "labels": { "app": "nginx" } }, "spec": { "containers": [ { "image": "nginx:1.14.2", "name": "nginx" } ] } } } } }
2023-06-25T08:11:18Z	INFO	sync-controller	predicate success	{"gvk": "apps/v1, Kind=Deployment", "key": {"name":"nginx","namespace":"default"}}
2023-06-25T08:11:18Z	INFO	sync-controller	reconciling instance	{"namespace": "default", "name": "nginx"}
```

6. Publish message deletion from another terminal:

```shell
mosquitto_pub -h 127.0.0.1 -p 31320 -u admin -P password -t "v1/client01/testing001/content" -m '{ "sentTimestamp": 1680964785, "resourceGenerationID": "12345", "content": { "apiVersion": "apps/v1", "kind": "Deployment", "metadata": { "uid": "123", "name": "nginx", "namespace": "default", "generation": 24, "deletionTimestamp": "2023-04-07T14:50:44Z" } } }'
```

7. Check the logs from example controller:

```shell
...
2023-06-25T08:22:51Z	INFO	mqtt-connection	received: v1/client01/testing001/content { "sentTimestamp": 1680964785, "resourceGenerationID": "12345", "content": { "apiVersion": "apps/v1", "kind": "Deployment", "metadata": { "uid": "123", "name": "nginx", "namespace": "default", "generation": 24, "deletionTimestamp": "2023-04-07T14:50:44Z" } } }
2023-06-25T08:22:51Z	INFO	sync-controller	predicate success	{"gvk": "apps/v1, Kind=Deployment", "key": {"name":"nginx","namespace":"default"}}
2023-06-25T08:22:51Z	INFO	sync-controller	reconciling instance	{"namespace": "default", "name": "nginx"}
2023-06-25T08:22:51Z	INFO	sync-controller	deleting	{"resource": "apps/v1, Kind=Deployment", "namespace": "default", "name": "nginx"}
```
