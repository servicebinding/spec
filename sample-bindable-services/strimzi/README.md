# Binding Event-driven Applications to an In-cluster Operator Managed Kafka Cluster

## Introduction

This scenario illustrates binding two event-driven applications to an in-cluster Operator Managed Kafka cluster using Service Binding Operator. OLM metadata including binding information is added to the Strimzi Operator to make the operator bindable via the Service Binding Operator.

This scenario also shows the use of the `customEnvVar` feature of the Service Binding Operator to specify a mapping for the injected environment variables.

## Actions to Perform by Users in 2 Roles

In this example there are 2 roles:

* Cluster Admin - Installs the operators to the cluster
* Application Developer - Creates a Kafka cluster, creates required Kafka resources, imports applications and creates a request to bind the applications to the Kafka cluster and the Kafka resources.

### Cluster Admin

The cluster admin needs to install 2 operators into the cluster:

* Service Binding Operator
* Strimzi Operator
* Runtime Component Operator

The [Strimzi Kafka Operator](https://github.com/navidsh/strimzi-kafka-operator) includes service binding metadata suggested by the [Service Binding Operator](https://github.com/redhat-developer/service-binding-operator/blob/master/docs/OperatorBestPractices.md) in order to be "bindable". The metadata is added to the operator's `ClusterServiceVersion` and exposes binding information through `secrets`, `status` fields and `spec` parameters.

#### Install the Service Binding Operator

Navigate to the **Operator** → **OperatorHub** in the OpenShift console. From the `Developer Tools` category, select the `Service Binding Operator` operator and install a `alpha` version. This would install the `ServiceBindingRequest` *custom resource definition* (CRD) in the cluster.

#### Install the Strimzi Operator

Create the following `OperatorSource`:

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

Once the `OperatorSource` is created, the "bindable" Strimzi Operator will be available to install from OperatorHub catalog in OpenShift console.

Then, navigate to the **Operator** → **OperatorHub** in the OpenShift console. Select **Strimzi** operator labelled as "custom" (not "Community") and install on the cluster.

#### Install the Runtime Component Operator

Navigate to the **Operator** → **OperatorHub** in the OpenShift console. Select the `Runtime Component Operator` operator and install a `beta` version. This would install the `RuntimeComponent` *custom resource definition* (CRD) in the cluster.

### Application Developer

#### Create a namespace called `service-binding-demo`

The application and the service instance needs a namespace to live in so let's create one for them:

```console
cat <<EOS |kubectl apply -f -
---
kind: Namespace
apiVersion: v1
metadata:
  name: service-binding-demo
EOS
```

#### Create a Kafka cluster

Now we use the Strimzi Operator that the cluster admin has installed. To create a Kafka cluster instance, just create a [`Kafka` custom resource](./01-kafka.yaml) in the `service-binding-demo` namespace called `my-cluster`:

```console
oc apply -f 01-kafka.yaml
```

It takes usually a few seconds to spin-up a new Kafka cluster. You can check the status in the custom resources:

```console
oc get kafka my-cluster -o yaml
```

This creates a Kafka cluster with a `tls` listener with mutual TLS authentication.

#### Create a Kafka Topic

To create a Kafka Topic, apply the [`KafkaTopic` custom resource](./02-kafka-topic.yaml) in the `service-binding-demo` namespace:

```console
oc apply -f 02-kafka-topic.yaml
```

Check the created custom resource:

```console
oc get kt my-topic
```

#### Create Kafka Users

To create Kafka Users, one user for producer application and one user for consumer application, apply the [`KafkaUser` custom resources](./03-kafka-user.yaml) in the `service-binding-demo` namespace:

```console
oc apply -f 03-kafka-user.yaml
```

Check the created custom resource:

```console
oc get kafkausers
```

#### Create a Service Binding

In order to bind the application to the Kafka resources, create [`ServiceBindingRequest` custom resources](./04-service-binding.yaml) in the `service-binding-demo` namespace:

```console
oc apply -f 04-service-binding.yaml
```

This would create two `ServiceBindingRequest`s. The custom resource for the producer application looks as follows:

```yaml
apiVersion: apps.openshift.io/v1alpha1
kind: ServiceBindingRequest
metadata:
  name: kafka-producer-binding
spec:
  backingServiceSelectors:
    - group: kafka.strimzi.io
      kind: KafkaUser
      resourceRef: my-producer
      version: v1beta1
    - group: kafka.strimzi.io
      kind: Kafka
      resourceRef: my-cluster
      version: v1beta1
  customEnvVar:
    - name: BOOTSTRAP_SERVERS
      value: |-
        {{- range .status.listeners -}}
          {{- if and (eq .type "tls") (gt (len .addresses) 0) -}}
            {{- with (index .addresses 0) -}}
              {{ .host }}:{{ .port }}
            {{- end -}}
          {{- end -}}
        {{- end -}}
```

There are two interesting parts in the binding requests:

* `backingServiceSelector` - used to specify the backing services. Applications need binding information from `Kafka` and `KafkaUser` resources.
* `customEnvVar` - specifies the mapping for the environment variables injected into the binding secret.

When the `ServiceBindingRequest` is created the Service Binding Operator's controller collects binding information from the referenced backing services and the `customEnvVar` and then stores into an intermediate Secret called with the same name as the `ServiceBindingRequest`. The `Deployment` resources have pointers to this secret to inject environment variables into the application containers.

#### Deploy the Producer and the Consumer Applications

To deploy the producer and the consumer applications, create [`RuntimeComponent` resources](./05-runtime-component.yaml) in the `service-binding-demo` namespace:

```console
oc apply -f 05-runtime-component.yaml
```

The Runtime Component Operator will create `Deployment` resources and defines binding information as environment variables for the application containers with values from the binding secret created by the Service Binding Operator.

#### View Messages

To see messages sent by the producer application, run:

```console
oc logs -n service-binding-demo -f $(oc get pods -l app=kafka-producer -o name)
```

You can press <kbd>Ctrl</kbd>+<kbd>C</kbd> to stop viewing the log messages.

Similarly, use the following command to see messages received by the consumer application:

```console
oc logs -n service-binding-demo -f $(oc get pods -l app=kafka-consumer -o name)
```

---

#### Further discussion on Service Binding Requests

To allow the service binding operator to automatically gather binding information such as the Kafka listener and certificates, some improvements would be required to the Service Binding Operator and/or the Strimzi operator. A [proposal for service binding support in Strimzi](https://github.com/strimzi/strimzi-kafka-operator/pull/2753) is currently being discussed.

These are illustrated in 2 steps:

### Step 1 - Improvements only to the Service Binding Operator

[service-binding-step1.yaml](./service-binding-step1.yaml) demonstrates what is possible with changes to only the Service Binding Operator. In this case, we bind directly to the secrets created by the Strimzi cluster and user operators to inject the CA certificate, user key and certificate. This means we depend on the names of these secrets, which is an implementation detail of the Strimzi operator. We also have a complicated Go template expression in order to retrieve the listener from the status attribute of the Kafka CR.

This would require the following changes to the Service Binding Operator:

* [Binding directly to secrets](https://github.com/redhat-developer/service-binding-operator/issues/389)
* [Refering to multiple backing services by ID](https://github.com/redhat-developer/service-binding-operator/issues/396)

### Stage 2 - Improvements to both Strimzi and the Service Binding Operator

With some changes to Strimzi, the service binding request is a lot nicer, as show in [service-binding-step2.yaml](service-binding-step2.yaml). We simply have to add the Kafka and KafkaUser CRs as backing services and the Service Binding Operator does all the hard work for us.

This would require the following changes to the Service Binding Operator:

* [Binding to complex objects in CRs](https://github.com/redhat-developer/service-binding-operator/issues/361)

On the Strimzi Operator side, the [proposal for service binding support in Strimzi](https://github.com/strimzi/strimzi-kafka-operator/pull/2753) lists changes needed to be done.
