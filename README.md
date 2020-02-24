# Service Binding Specification

Specification for binding services to runtime applications running in Kubernetes.  

## Terminology definition

*  **service** - any software that is exposing functionality.  Could be a RESTful application, a database, an event stream, etc.
*  **application** - in this specification we refer to a single runtime-based microservice (e.g. MicroProfile app, or Node Express app) as an application.  This is different than an umbrella (SIG) _Application_ which refers to a set of microservices.
*  **binding** - providing the necessary information for an application to connect to a service.
*  **Secret** - refers to a Kubernetes [Secret](https://kubernetes.io/docs/concepts/configuration/secret/).
*  **ConfigMap** - refers to a Kubernetes [ConfigMap](https://kubernetes.io/docs/tasks/configure-pod-container/configure-pod-configmap/).

## Motivation

*  Need a consistent way to bind k8s application to services (applications, databases, event streams, etc)
*  A standard / spec / RFC will enable adoption from different service providers
*  Cloud Foundry has done this well.  The equivalent is not available for k8s.

## Proposal

Main section of the doc.  Has sub-section that outline the design.

### 1.  Making a service bindable

#### Recommended
For a service to be bindable it **should** provide:
* a ConfigMap that contains the name of the Secret holding the binding data (see the `Service Binding Schema` section below), and describes metadata associated with each of the items referenced in the Secret, as well as a reference to this ConfigMap.

This pattern ensures the binding data is properly secured in a Secret and has corresponding metadata to enhance the consumption experience of the binding. 

The key/value pairs insides this ConfigMap are:
* A single `Secret=<name_of_secret>` - where `<name_of_secret>` is the qualified name (including the namespace) of the k8s Secret, which contains the binding data for this service.
* A set of `Metadata.<property>` - where `<property>` maps to one of the defined keys for this service, and `<value>` represents the description of the value.  For example, this is useful to define what format the `password` key is in, such as apiKey, basic auth password, token, etc.


#### Minimum
If not following the recommended path above, for a service to be bindable it **must** provide:
* a Secret that contains the binding data and a reference to this Secret, OR, map its `status` properties to the corresponding binding data.

#### Pointer to binding data

The reference's location and format depends on the following scenarios:

1. OLM-enabled Operator: Use the `statusDescriptor` part of the CSV to mark which `status` properties reference the binding data:
    * The reference's `x-descriptors` with one-of:
      * ConfigMap:
        * servicebinding:ConfigMap
      * Secret:
        * servicebinding:Secret
      * Individual binding items:
        * servicebinding:Secret:host
        * servicebinding:Secret:port
        * servicebinding:Secret:uri
        * servicebinding:Secret:`<binding_property>`  (where `<binding_property>` is any property from the binding schema)

2. Non-OLM Operator: - An annotation in the Operator's CRD to mark which `status` properties reference the binding data:
    * The annotation is one-of:
      * ConfigMap:
        * servicebinding/configMap: status.bindable.ConfigMap
      * Secret:
        * servicebinding/secret: status.bindable.Secret
      * Individual binding items:
        * servicebinding/secret/host: status.address
        * servicebinding/secret/`<binding_property>`: status.`<status_property>` (where `<binding_property>` is any property from the binding schema, and `<status_property>` refers to a `status` property)

3. Regular k8s Deployment (Ingress, Route, Service, etc)  - An annotation in the corresponding CR that maps the `status` properties to their corresponding binding data:
      * servicebinding/secret/host: status.ingress.host
      * servicebinding/secret/host: status.address
      * servicebinding/secret/`<binding_property>`: status.`<status_property>` (where `<binding_property>` is any property from the binding schema, and `<status_property>` refers to a `status` property)

4. External service - An annotation in the local ConfigMap or Secret that bridges the external service.
    * The annotation is in the form of either:
      * servicebinding/configMap: self
      * servicebinding/secret: self

### 2.  Service Binding Schema

The core set of binding data is:
* **name** - name of the service.
* **host** - host (IP or host name) where the service resides.
* **port** - the port to access the service.
* **protocol** - protocol of the service.  Examples: http, https, postgre, mysql, mongodb, amqp, mqtt, etc.
* **username** - username to log into the service.  Can be omitted if no authorization required, or if equivalent information is provided in the password as a token.
* **password** - the password or token used to log into the service.  Can be omitted if no authorization required, or take another format such as an API key.  It is strongly recommended that the corresponding ConfigMap metadata properly describes this key.
* **certificate** - the certificate used by the client to connect to the service.  Can be omitted if no certificate is required, or simply point to another Secret that holds the client certificate.  
* **uri** - for convenience, the full URI of the service in the form of `<protocol>://<host>:<port>/<name>`.

Extra binding properties can also be defined (with corresponding metadata) in the bindable service's ConfigMap (or Secret).  For example, services may have credentials that are the same for any user (global setting) in addition to per-user credentials.


### 3.  Request service binding

* How do we request a binding from a service (assume the service has been provisioned)
  * One option:
    * apiVersion: apps.openshift.io/v1alpha1
    * kind: ServiceBindingRequest

* How is that binding authorized?

### 4.  Mounting binding information

* Where in the container do we mount the binding information (e.g. what is the structure of the folders / files)
  * One option:
    * `platform/bindings/<service-id>/metadata`
    * `platform/bindings/<service-id>/secret`

* Consideration with clusters, namespaces, or VMs

### Extra:  Consuming binding

*  How are application expected to consume binding information 
*  Each framework may take a different approach, so this is about samples & recommendations (best practices)
*  Validates the design
