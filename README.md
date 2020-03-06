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
*  Cloud Foundry has addressed this issue with a [buildpack specification](https://github.com/buildpacks/rfcs/blob/master/text/0012-service-binding.md). The equivalent is not available for k8s.

## Proposal

Main section of the doc.  Has sub-sections that outline the design.

### Making a service bindable

#### Minimum
For a service to be bindable it **MUST** comply with one-of:
* provide a Secret that contains the binding data and reference this Secret using one of the patterns discussed [below](#pointer-to-binding-data). 
* map its `status` properties to the corresponding binding data, using one of the patterns discussed [below](#pointer-to-binding-data).
* include a sample `ServiceRequestBinding` (see the [Request service binding](#Request-service-binding) section below) in its documentation (e.g. GitHub repository, installation instructions, etc) which contains a `dataMapping` illustrating how each of its `status` properties map to the corresponding binding data.  This option allows existing services to be bindable with zero code changes.

#### Recommended
In addition to the minimum set above, a bindable service **SHOULD** provide:
* a ConfigMap that describes metadata associated with each of the items referenced in the Secret.  The bindable service should also provide a reference to this ConfigMap using one of the patterns discussed [below](#pointer-to-binding-data).

The key/value pairs insides this ConfigMap are:
* A set of `Metadata.<property>=<value>` - where `<property>` maps to one of the defined keys for this service, and `<value>` represents the description of the value.  For example, this is useful to define what format the `password` key is in, such as apiKey, basic auth password, token, etc.


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

2. Non-OLM Operator: - An annotation in the Operator's CRD to mark which `status` properties reference the binding data.  The value of this annotation can be specified in either [JSONPath](https://kubernetes.io/docs/reference/kubectl/jsonpath/) or [GO templates](https://golang.org/pkg/text/template/):
      * ConfigMap:
        * servicebinding/configMap: {.status.bindable.ConfigMap}
      * Secret:
        * servicebinding/secret: {.status.bindable.Secret}
      * Individual binding items:
        * servicebinding/secret/host: {.status.address}
        * servicebinding/secret/`<binding_property>`: {.status.`<status_property>}` (where `<binding_property>` is any property from the binding schema, and `<status_property>` refers to the path to the correspoding `status` property)

3. Regular k8s Deployment (Ingress, Route, Service, etc)  - An annotation in the corresponding CR that maps the `status` properties to their corresponding binding data. The value of this annotation can be specified in either [JSONPath](https://kubernetes.io/docs/reference/kubectl/jsonpath/) or [GO templates](https://golang.org/pkg/text/template/):
      * servicebinding/secret/host: {.status.ingress.host}
      * servicebinding/secret/host: {.status.address}
      * servicebinding/secret/`<binding_property>`: status.`<status_property>` (where `<binding_property>` is any property from the binding schema, and `<status_property>` refers to the path to the correspoding `status` property)

4. External service - An annotation in the local ConfigMap or Secret that bridges the external service.
    * The annotation is in the form of either:
      * servicebinding/configMap: self
      * servicebinding/secret: self

### Service Binding Schema

The core set of binding data is:
* **type** - the type of the service. Examples: openapi, db2, kafka, etc.
* **host** - host (IP or host name) where the service resides.
* **port** - the port to access the service.
* **protocol** - protocol of the service.  Examples: http, https, postgresql, mysql, mongodb, amqp, mqtt, etc.
* **username** - username to log into the service.  Can be omitted if no authorization required, or if equivalent information is provided in the password as a token.
* **password** - the password or token used to log into the service.  Can be omitted if no authorization required, or take another format such as an API key.  It is strongly recommended that the corresponding ConfigMap metadata properly describes this key.
* **certificate** - the certificate used by the client to connect to the service.  Can be omitted if no certificate is required, or simply point to another Secret that holds the client certificate.  
* **uri** - for convenience, the full URI of the service in the form of `<protocol>://<host>:<port>/<name>`.
* **role-needed** - the name of the role needed to fetch the Secret containing the binding data.  In this scenario a Service Account with the appropriate role must be passed into the binding request (see the [RBAC](#rbac) section below).

Extra binding properties can also be defined (with corresponding metadata) in the bindable service's ConfigMap (or Secret).  For example, services may have credentials that are the same for any user (global setting) in addition to per-user credentials.


### Request service binding

Binding is requested by the consuming application, or an entity on its behalf such as the [Runtime Component Operator](https://github.com/application-stacks/runtime-component-operator), via a custom resource that is applied in the same cluster where an implementation of this specification resides.

Since the reference implementation for most of this specification is the [Service Binding Operator](https://github.com/redhat-developer/service-binding-operator) we will be using the `ServiceBindingRequest` CRD, which resides in [this folder](https://github.com/redhat-developer/service-binding-operator/tree/master/deploy/crds).  

**Temporary Note**
To ensure a better fit with the specification a few modifications have been proposed to the `ServiceBindingRequest` CRD:
* A modification to its API group, to be independent from OpenShift.
* A simplification of its CRD name to `ServiceBinding`.
* Renaming `customEnvVar` to `dataMapping`.
* Allowing for the `application` selector to be omitted, for the cases where another Operator owns the deployment.
* Addition of fields such as `serviceAccount` and `subscriptionSecret` that support more advanced binding cases (more below).

#### RBAC

If the service provider's Secret (as defined in [Pointer to binding data](#pointer-to-binding-data)) is protected by RBAC then the service consumer must pass a Service Account in its `ServiceBindingRequest` CR to allow implementations of this specification to fetch information from that Secret.  If the implementation already has access to all Secrets in the cluster (as is the case with the Service Binding Operator) it must ensure it uses the provided Service Account instead of its own - blocking the bind if a Service Account was needed (accordingin to the binding data) by not provided.

Example of a partial CR:

```
 services:
    - group: postgres.dev
      kind: Service
      resourceRef: user-db
      version: v1beta1
      serviceAccount: <my-sa>
```

#### Subscription-based services

There are a variety of service providers that require a subscription to be created before accessing the service. Examples:
* an API management framework that provides apiKeys after a plan subscription has been approved
* a database provisioner that spins single-tenant databases upon request
* premium services that deploy a providers located in the same physical node as the caller for low latency
* any other type of subscription service

The only requirement from this specification is that the subscription results in a Secret, containing a partial or complete set of binding data (defined in [Service Binding Schema](#service-binding-schema)), created in the same namespace of the `ServiceBindingRequest` that will reference this Secret.  Implementations of this specification will populate this subscription Secret with any additional information provided by the target service, according to the [Pointer to binding data](#pointer-to-binding-data) section.

Example of a partial CR:

```
 services:
    - group: postgres.dev
      kind: Service
      resourceRef: user-db
      version: v1beta1
      subscriptionSecret: <my-subscription-secret>
```


### Mounting binding information

Implementations of this specification must bind the Secret containing binding data using the following format:

```
<path>/bindings/<service-id>/metadata/configMap/<persisted_configMap>
<path>/bindings/<service-id>/metadata/request/<ServiceBindingData_CR>
<path>/bindings/<service-id>/secret/<persisted_secret>
```

Where:
* `<path>` defaults to `platform` if not specified in the `ServiceBindingRequest` CR.
* `<service-id>` equals the `metadata.name` field from the `ServiceBindingRequest` CR.
* `<persisted_configMap>` represents a set of files where the filename is a ConfigMap key and the file contents is the corresponding value of that key.
* `<ServiceBindingData_CR>` represents the requested `ServiceBindingRequest` CR
* `<persisted_secret>` represents a set of files where the filename is a Secret key and the file contents is the corresponding value of that key.


### Extra:  Consuming binding

*  How are application expected to consume binding information 
*  Each framework may take a different approach, so this is about samples & recommendations (best practices)
*  Validates the design
