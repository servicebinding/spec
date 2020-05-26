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

* Need a consistent way to bind k8s application to services (applications, databases, event streams, etc).
* A standard / spec / RFC will enable adoption from different service providers.
* The CNB (Cloud Native Buildpacks) community has started to address this issue with a [bindings specification](https://github.com/buildpacks/spec/blob/master/extensions/bindings.md). We should align with these concepts and expand the use case to incorporate scenarios such as: 
  * how to make a service bindable
  * how to request a binding
  * what is the proposed binding schema

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

Application **SHOULD** rely on the `SERVICE_BINDINGS` environment variable for the accurate list of injected or mounted binding items, as [defined below](#Mounting-and-injecting-binding-information).


### Request service binding

Binding is requested by the consuming application, or an entity on its behalf such as the [Runtime Component Operator](https://github.com/application-stacks/runtime-component-operator), via a custom resource that is applied in the same cluster where an implementation of this specification resides.

Since the reference implementation for most of this specification is the [Service Binding Operator](https://github.com/redhat-developer/service-binding-operator) we will be using the `ServiceBinding` CRD, which resides in [this folder](https://github.com/redhat-developer/service-binding-operator/tree/master/deploy/crds), as the entity that holds the binding request.  

**Note** - a few updates / enhancements are being proposed to the current `ServiceBinding` CR, tracked [here](https://github.com/application-stacks/service-binding-specification/issues/16#issuecomment-605309629).


#### Subscription-based services

There are a variety of service providers that require a subscription to be created before accessing the service. Examples:
* an API management framework that provides apiKeys after a plan subscription has been approved
* a database provisioner that spins single-tenant databases upon request
* premium services that deploy a providers located in the same physical node as the caller for low latency
* any other type of subscription service

The only requirement from this specification is that the subscription results in a k8s resources (Secret, etc), containing a partial or complete set of binding data (defined in [Service Binding Schema](#service-binding-schema)).  From the `ServiceBinding` CR's perspective, this resource looks and feels like an additional service.

Example of a partial CR:

```
 services:
    - group: postgres.dev
      kind: Service
      resourceRef: global-user-db
      version: v1beta1
    - group: postgres.dev
      kind: Secret
      resourceRef: specific-user-db
      version: v1beta1      
```


### Mounting and injecting binding information

This specification allows for data to be mounted using volumes or injected using environment variables.  The best practice is to mount any sensitive information, such as passwords, since that will avoid accidentally exposure via environment dumps and subprocesses.  Also, binding binary data (e.g. .p12 certificate for Kafka) as an environment variable might cause a pod to fail to start (stuck on `CrashLoopBackOff`), so it advisable for backing services with such binding data to mark it with `bindAs: volume`.

The decision to mount vs inject is made in the following ascending order of precedence:
* value of the `bindAs` attribute in the backing service as defined in its [annotations](annotations.md#data-model--building-blocks-for-expressing-binding-information), applying to the binding item referenced by the annotation.
* value of `ServiceBinding`'s global `bindAs` element, which applies to all binding data.
* value of the `bindAs` attribute in each of the `dataMappings` elements inside `ServiceBinding`.

#### Injecting data

The key `SERVICE_BINDINGS` acts as a global map of the service bindings and **MUST** always be injected into the environment.  It contains a JSON payload with `bindingKeys` key name containing a list of all available binding information available. Each item of the `bindingKeys` list includes an object containing `name`, `bindAs` and an optional `mountPath` (if it is bound as a volume).

Example:

```json
SERVICE_BINDINGS = {
  "bindingKeys": [
    {
      "name": "KAFKA_USERNAME",
      "bindAs": "envVar"
    },
    {
      "name": "KAFKA_PASSWORD",
      "bindAs": "volume",
      "mountPath": "/platform/bindings/secret/"
    }
  ]
}
```

In the example above, the application **MAY** query the environment variable `SERVICE_BINDINGS`, walk its JSON payload and learn that `KAFKA_USERNAME` is available as an environment variable, and that `KAFKA_PASSWORD` is available as a mounted file inside the directory `/platform/bindings/secret/`.

#### Mounting data
Implementations of this specification must bind the following data into the consuming application container:

```
<mountPathPrefix>/bindings/secret/<persisted_secret>
```

Where:
* `<mountPathPrefix>` defaults to `platform` if not specified in the `ServiceBinding` CR via the `mountPathPrefix` element.
* `<persisted_secret>` represents a set of files where the filename is a Secret key and the file contents is the corresponding value of that key.
