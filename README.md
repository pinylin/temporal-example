
# env

## k8s
```shell
Kubernetes control plane is running at https://127.0.0.1:59537
CoreDNS is running at https://127.0.0.1:59537/api/v1/namespaces/kube-system/services/kube-dns:dns/proxy

kubectl proxy
127.0.0.1:8001

```

## temporal


```
git clone
cd helm-charts/charts/temporal
git reset --hard 1e5ac0c

helm install \                                          
    --set server.replicaCount=1 \
    --set cassandra.config.cluster_size=1 \
    --set prometheus.enabled=false \
    --set grafana.enabled=false \
    --set elasticsearch.enabled=false \
    temporaltest . --timeout 15m

```
```
kubectl exec -it services/temporaltest-admintools /bin/bash

kubectl port-forward services/temporaltest-frontend-headless 7233:7233
kubectl port-forward services/temporaltest-web 8080:8080
```

## kafka

helm install kafka oci://registry-1.docker.io/bitnamicharts/kafka

```yaml

# docker run  -d --name kafka -p 9092:9092 -e KAFKA_BROKER_ID=0 -e KAFKA_ZOOKEEPER_CONNECT=://10.0.4.13:2181 -e KAFKA_ADVERTISED_LISTENERS=PLAINTEXT://10.0.4.13:9092 -e KAFKA_LISTENERS=PLAINTEXT://0.0.0.0:9092 -t wurstmeister/kafka

docker run -d --name kafka-test -p 9092:9092 \
--link zookeeper \
--env KAFKA_ZOOKEEPER_CONNECT=zookeeper:2181 \
--env KAFKA_ADVERTISED_HOST_NAME=localhost \
--env KAFKA_ADVERTISED_PORT=9092  \
wurstmeister/kafka
```

## dockefile
```
$ docker build -t temporal-example:v0.01 .

$ docker tag temporal-example:v0.01 localhost:5001/temporal-example:v0.01

$ docker push localhost:5001/temporal-example:v0.01
```
# debug
## topic
```
1.
kubectl exec --stdin --tty kafka-controller-0 -- /bin/bash
/opt/bitnami/kafka/bin/kafka-topics.sh --create --topic topic1 --bootstrap-server kafka:9092 
echo "topic topic1 was create"


2. 还是不行
kubectl run kafka-producer -ti \
--image=quay.io/strimzi/kafka:0.40.0-kafka-3.7.0 --rm=true --restart=Never \
-- bin/kafka-console-producer.sh --bootstrap-server 10.96.36.133:9092 --topic topic1

3. 直接代码
```
# questions

1. ERROR Unable to start temporal worker !BADKEY="Namespace default is not found.
> tctl --ns default namespace register -rd 3


2. yaml 配置不生效, 所以是判断pod_name suffix, 应该有更合适/优雅的方式
```yaml
    - name: IS_MASTER
      value: "{{ eq .POD_NAME \"temporal-example-ss-0\" }}"
```

3. "Error":"unable to find workflow type: CronConsumerWorkflow. Supported types: [CronProducerWorkflow, ChildCronProducerWorkflow]"}

# todo
- [ ] helm 