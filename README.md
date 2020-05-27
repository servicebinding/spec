# Service Binding Specification

Specification for binding services to runtime applications running in Kubernetes.  

## Conventions

The key words "MUST", "MUST NOT", "REQUIRED", "SHALL", "SHALL NOT", "SHOULD",
"SHOULD NOT", "RECOMMENDED", "NOT RECOMMENDED", "MAY", and "OPTIONAL" in this
document are to be interpreted as described in [BCP
14](https://tools.ietf.org/html/bcp14)
[[RFC2119](https://tools.ietf.org/html/rfc2119)]
[[RFC8174](https://tools.ietf.org/html/rfc8174)] when, and only when, they
appear in all capitals, as shown here.

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

The following sections outline the details of the specification.

### Making a service bindable

#### Minimum requirements for being bindable
A bindable service **MUST** comply with one-of:
* provide a Secret and/or ConfigMap that contains the [binding data](#service-binding-schema) and reference this Secret and/or ConfigMap using one of the patterns discussed [below](#pointer-to-binding-data). 
* map its `status`, `spec`, `data` properties to the corresponding [binding data](#service-binding-schema), using one of the patterns discussed [below](#pointer-to-binding-data).
* include a sample `ServiceBinding` (see the [Request service binding](#request-service-binding) section below) in its documentation (e.g. GitHub repository, installation instructions, etc) which contains either:
  * a `dataMapping` element illustrating how each of its `status`, `spec` or `data` properties map to the corresponding [binding data](#service-binding-schema).  
  * a `detectBindingResources: true` element which will automatically populate the resulting Secret from the `ServiceBinding` with information from any Route, Ingress, Service, ConfigMap or Secret resources that are owned by the backing service CR.

<kbd>EXPERIMENTAL</kbd>The service **MUST** also make itself discoverable by complying with one-of:
* In the case of an OLM-based Operator, add `Bindable` to the CSV's `metadata.annotations.categories`.
* In the case of a Helm chart service, add bindable to the Chart.yaml's keyword list.
* In all other cases, add the `servicebinding/bindable: "true"` annotation to your CRD or any CR (Secret, Service, etc).

#### Pointer to binding data

This specification supports different scenarios for exposing bindable data. Below is a summary of how to indicate what is interesting for binding.  Please see the [annotation section](annotations.md) for the full set with more details.

1. OLM-enabled Operator: Use the `statusDescriptor` and/or `specDescriptor` parts of the CSV to mark which `status` and/or `spec` properties reference the [binding data](#service-binding-schema):
    * The reference's `x-descriptors` with a possible combination of:
      * ConfigMap:
        
            - path: data.dbcredentials
              x-descriptors:
                - urn:alm:descriptor:io.kubernetes:ConfigMap 
                - servicebinding

      * Secret:

            - path: data.dbcredentials
              x-descriptors:
                - urn:alm:descriptor:io.kubernetes:Secret 
                - servicebinding        

      * Individual binding items from a `Secret`:
      
            - urn:alm:descriptor:io.kubernetes:Secret 
            - servicebinding:username
          
            - urn:alm:descriptor:io.kubernetes:Secret 
            - servicebinding:password
          
      * Individual binding items from a `ConfigMap`:
      
            - urn:alm:descriptor:io.kubernetes:ConfigMap 
            - servicebinding:port

            - urn:alm:descriptor:io.kubernetes:ConfigMap
            - servicebinding:host
            
      * Individual backing items from a path referencing a string value
      
            - path: data.uri
              x-descriptors:
                - servicebinding 
          
2. Non-OLM Operator: - An annotation in the Operator's CRD to mark which `status` and/or `spec` properties reference the [binding data](#service-binding-schema) :
      * ConfigMap:
      
            "servicebinding.dev/certificate":
            "path={.status.data.dbConfiguration},objectType=ConfigMap"
     
      * Secret:
      
            "servicebinding.dev/dbCredentials":
            "path={.status.data.dbCredentials},objectType=Secret"

      * Individual binding items from a `ConfigMap`

            “servicebinding.dev/host": 
             “path={.status.data.dbConfiguration},objectType=ConfigMap,sourceKey=address"

            “servicebinding.dev/port": 
            “path={.status.data.dbConfiguration},objectType=ConfigMap

      * Individual backing items from a path referencing a string value

            “servicebinding.dev/uri”:"path={.status.data.connectionURL}"
      
3. Regular k8s resources (Ingress, Route, Service, Secret, ConfigMap etc)  - An annotation in the corresponding Kubernetes resources that maps the `status`, `spec` or `data` properties to their corresponding [binding data](#service-binding-schema). 

All annotations used in CRDs in the above section **MAY** be used for regular k8s resources, as well.

The above pattern **MAY** be used to expose external services (such as from a VM or external cluster), as long as there is an entity such as a Secret that provides the binding details.

### Service Binding Schema

The core set of binding data is:
* **type** - the type of the service. Examples: openapi, db2, kafka, etc.
* **host** - the host (IP or host name) where the service resides.
* **port** - the port to access the service.
* **endpoints** - the endpoint information to access the service in a service-specific syntax, such as a list of hosts and ports. This is an alternative to separate `host` and `port` properties.
* **protocol** - the protocol of the service.  Examples: http, https, postgresql, mysql, mongodb, amqp, mqtt, kafka, etc.
* **basePath** - a URL prefix for this service, relative to the host root. It MUST start with a leading slash `/`.  Example: the URL prefix for a RESTful service. 
* **username** - the username to log into the service.  **MAY** be omitted if no authorization required, or if equivalent information is provided in the password as a token.
* **password** - the password or token used to log into the service.  **MAY** be omitted if no authorization required, or take another format such as an API key.  It is strongly recommended that the corresponding ConfigMap metadata properly describes this key.
* **certificate** - the certificate used by the client to connect to the service.  **MAY** be omitted if no certificate is required, or simply point to another Secret that holds the client certificate.
* **uri** - for convenience, the full URI of the service in the form of `<protocol>://<host>:<port>[<basePath>]`.
* <kbd>EXPERIMENTAL</kbd>**bindingRole** - the name of the role needed to read the binding data exposed from this bindable service.  Implementations of this specification **SHOULD** enforce that only those with the appropriate roles are allowed to access the binding data.

Extra binding properties **SHOULD** also be defined by the bindable service, using one of the patterns defined in [Pointer to binding data](#pointer-to-binding-data).

The data that is injected or mounted into the container may have a different name because of a few reasons:
* the backing service may have chosen a different name (discouraged, but allowed).
* a custom name may have been chosen using the `dataMappings` portion of the `ServiceBinding` CR.
* a prefix may have been added to certain items in `ServiceBinding`, either globally or per service.  

### Request service binding

Binding is requested by the consuming application, or an entity on its behalf such as the [Runtime Component Operator](https://github.com/application-stacks/runtime-component-operator), via a custom resource that is applied in the same cluster where an implementation of this specification resides.

Since the reference implementation for most of this specification is the [Service Binding Operator](https://github.com/redhat-developer/service-binding-operator) we will be using the `ServiceBinding` CRD, which resides in [this folder](https://github.com/redhat-developer/service-binding-operator/tree/master/deploy/crds), as the entity that holds the binding request.  

**Note** - a few updates / enhancements are being proposed to the current `ServiceBinding` CR, tracked [here](https://github.com/application-stacks/service-binding-specification/issues/16#issuecomment-605309629).

Sample CR
```
apiVersion: service.binding/v1alpha1
kind: ServiceBinding
metadata:
  name: example-service-binding
spec:
  mountPathPrefix: "/var/bindings"
  services:
    - group: postgres.dev
      kind: Service
      name: global-user-db
      version: v1beta1
      id: postgres-global-user
    - group: ibmcloud.ibm.com
      version: v1alpha1
      kind: Binding
      name: coligo-service-binding
      id: coligo-service-binding
  application:
    name: nodejs-rest-http-crud
    group: apps
    version: v1
    resource: deployments
```

#### Customizing data bindings

The `ServiceBinding` CR has an element per service called `dataMappings`, whereby the author of the CR can rename certain binding items and/or compose of multiple items:

Partial sample of service-specific mappings:
```
  services:
    - group: postgres.dev
      kind: Service
      name: global-user-db
      version: v1beta1
      id: postgres-global-user
    - group: ibmcloud.ibm.com
      version: v1alpha1
      kind: Binding
      name: watson-service-binding
      id: watson-service-binding
      dataMappings:
        - name: WATSON_URL
          value: {{  .status.host }} / {{ .status.port }}
        - name: WATSON_USERNAME
          value: {{  .status.username }}
```

#### <kbd>EXPERIMENTAL</kbd> Synthetic data bindings

If a `dataMappings` requires a cross-service composition then a new synthetic service entry must be created.  

Partial sample of a synthetic / composed mapping:
```
  services:
    - group: event.stream
      kind: User
      name: my-user
      version: v1beta1
      id: event-user
    - group: event.stream
      kind: Cluster
      resourceRef: my-cluster
      version: v1beta1
      id: event-cluster
    - group: servicebinding
      version: v1alpha1
      kind: ComposedService
      name: events-composed
      id: event-stream
      dataMappings:
        - name: EVENT_STREAMS_URL
          value: {{  event-user.status.url }} / ?username= {{ event-user.status.username }}
```

The service entry with apiVersion `servicebinding/v1alpha1` and kind `ComposedService` refers to a synthetic CR whose sole purpose is to compose bindings from other services.  

#### Subscription-based services

There are a variety of service providers that require a subscription to be created before accessing the service. Examples:
* an API management framework that provides apiKeys after a plan subscription has been approved
* a database provisioner that spins single-tenant databases upon request
* premium services that deploy a providers located in the same physical node as the caller for low latency
* any other type of subscription service

The only requirement from this specification is that the subscription results in a k8s resources (Secret, etc), containing a partial or complete set of binding data (defined in [Service Binding Schema](#service-binding-schema)).  From the `ServiceBinding` CR's perspective, this resource looks and feels like an additional service.

Example of a partial CR, where the second service refers to a Secret containing the provisioned user specific credentials.
```
 services:
    - group: postgres.dev
      kind: Service
      name: global-user-db
      version: v1beta1
    - group: postgres.dev
      kind: Secret
      name: specific-user-db
      version: v1beta1      
```


### Mounting and injecting binding information

#### Mounting data
Implementations of this specification **MUST** mount the binding data into the consuming application container in the following location:

```
$SERVICE_BINDINGS_ROOT/<service-id>/secret/<persisted_bindings>
```

Where:
* $SERVICE_BINDINGS_ROOT is an immutable environment variable (once set) that corresponds to the root location of the bindings.  This value comes from the first `ServiceBinding` CR, as there may be many, and its `mountPathPrefix` element, or from the default value of `/platform/bindings` if the first `ServiceBinding` CR did not specify a `mountPathPrefix` element.
  * This means that if another `ServiceBinding` CR wants to project itself into the same container, it **MUST** reuse the current `SERVICE_BINDINGS_ROOT` value, even if it had a conflicting `mountPathPrefix` element.
* `<service-id>` is the `id` field of the corresponding `service` entry in the `ServiceBinding` CR.  If the `id` field is not present, the `name` field is used instead.  The `<service-id>` path **MUST** be unique between the services bound to a particular application.
* `<persisted_secret>` represents a set of files where the filename is a Secret key and the file contents is the corresponding value of that key.

Example:  `/platform/bindings/my-nosql-db/secret/MONGODB_HOST`

In addition to the secret, implementations of this specification **MUST** mount, if available, metadata about the bindings in the following location:

```
$SERVICE_BINDINGS_ROOT/<service-id>/metadata/<persisted_metadata>
```

The recommended set of metadata will vary dependending on implementations and platforms, but two **RECOMMENDED** keys are:
* kind
* provider


#### Exposing data as environment variables

The specification allows for binding properties to be additionally exposed as environment variables.  The assumption is that applications will have pre-knowledge of these environment variable keys, so the mechanisms described in this section are solely for the purpose of allowing the entity responsible for mounting the binding data to know which keys should also be made available as environment variables.  

This decision is made in the following ascending order of precedence:
* value of `ServiceBinding`'s global `bindAs` element, which applies to all binding data.
* value of the `bindAs` attribute in each of the `dataMappings` elements inside `ServiceBinding`.
