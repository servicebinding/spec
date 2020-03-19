# Binding to Strimzi Operator

This sample demonstrates:
- deploying a Kafka cluster with a listener with mutual TLS authentication
- creating a Kafka topic
- creating a Kafka user with mutual TLS authentication
- deploying a producer and a consumer
- 🚧creating a `ServiceBindingRequest` to inject Kafka listener address and various certificates into applications

## Requirements

This sample expects users to execute against an OpenShift Container Platform 4.3 cluster. Also, it expects users to run commands in `myproject` namespace:

```
oc new-project myproject
```

## Install Strimzi Operator

Navigate in the OpenShift web console to the **Operators** → **OperatorHub** page. Install Strimzi Operator onto the cluster.

## Deploy Kafka cluster

Create a Kafka cluster:

```console
oc apply -f 01-kafka.yaml
```

Monitor the cluster deployment progress:

```console
oc get all
```

## Create `KafkaTopic` and `KafkaUser` resources

Create a Kafka Topic:

```console
oc apply -f 02-topic.yaml
```

Create a Kafka User:

```console
oc apply -f 03-users.yaml
```

## Kafka Producer and Consumer Application

Deploy a Kafka producer and a Kafka consumer:
```console
oc apply -f 04-deployments.yaml
```

Go through the file `04-deployments.yaml` and look at the different environment variables defined in `spec.template.spec.containers.env`.
