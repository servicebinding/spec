# Binding Event-driven Applications to an In-cluster Operator Managed Kafka Cluster

## Introduction

This scenario illustrates binding two event-driven applications to an in-cluster Operator Managed Kafka cluster using Service Binding Operator. OLM metadata including binding information is added to the Strimzi Operator to make the operator bindable via the Service Binding Operator.

This scenario also shows the use of the `customEnvVar` feature of the Service Binding Operator to specify a mapping for the injected environment variables.

## Requirements

This sample expects users to execute against an OpenShift Container Platform 4.3 cluster. Also, it expects users to run commands in `myproject` namespace:

```
oc new-project myproject
```

## Install the Strimzi Operator using an `OperatorSource`

Apply the following `OperatorSource`:

```console
cat <<EOS |kubectl apply -f -
---
apiVersion: operators.coreos.com/v1
kind: OperatorSource
metadata:
  name: strimzi-operators
  namespace: openshift-marketplace
spec:
  type: appregistry
  endpoint: https://quay.io/cnr
  registryNamespace: navidsh
  displayName: "Bindable Strimzi Operators"
EOS
```

Then navigate in the OpenShift web console to the **Operators** → **OperatorHub** page. Install the Strimzi Operator labeled as **Custom** onto the cluster.

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

Create a Kafka Topic and Kafka Users:

```console
oc apply -f 02-topic-users.yaml
```

## Express an intent to bind the Kafka cluster and the applications

Deploy a Kafka producer and a Kafka consumer:
```console
oc apply -f 04-deployments.yaml
```

---

## Further discussion on Service Binding Requests

To allow the service binding operator to automatically inject the Kafka listeneer and certificates (instead of manually doing so as in [04-deployments.yaml](04-deployments.yaml)), some improvements would be required to the Service Binding Operator and/or the Strimzi operator.

These are illustrated in 2 steps:

### Step 1 - Improvements only to the Service Binding Operator

[service-binding-step1.yaml](service-binding-step1.yaml) demonstrates what's possible with changes to only the Service Binding Operator. In this case, we bind directly to the secrets created by the Strimzi cluster and user operators to inject the CA certificate, user key and certificate. This means we depend on the names of these secrets, which is an implementation detail of the Strimzi operator.  We also have a complicated Go template expression in order to retrieve the listener from the status attribute of the Kafka CR.

This would require the following changes to the Service Binding Operator:
- [Binding directly to secrets](https://github.com/redhat-developer/service-binding-operator/issues/389)
- [Refering to multiple backing services by ID](https://github.com/redhat-developer/service-binding-operator/issues/396)

### Stage 2 - Improvements to both Strimzi and the Service Binding Operator

With some changes to Strimzi, the service binding request is a lot nicer, as show in [service-binding-step2.yaml](service-binding-step2.yaml).  We simply have to add the Kafka and KafkaUser CRs as backing services and the Service Binding Operator does all the hard work for us. 

This would require the following changes to the Service Binding Operator:
- [Binding to complex objects in CRs](https://github.com/redhat-developer/service-binding-operator/issues/361)

and the following changes to Strimzi:
- Exposing the listeners in a more useful way on the status of the Kafka CR. For example, the cluster operator could provide a property containing the concatentated list of listeners of each type.
- Exposing the secret containing the CA certficiate as an attribute of the status of the Kafka CR
- Adding the service binding annotations to the above attributes so the service binding operator can discover them.
- Adding the service binding annotations to the `secret` attribute of the KafkaUser CR so the service binding operator can discover it.
