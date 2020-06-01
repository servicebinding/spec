# Service Binding Specification for Kubernetes

Today in Kubernetes, the exposure of secrets for connecting applications to external services such as REST APIs, databases, event buses, and more is both manual and bespoke.  Each service provider suggests a different way to access their secrets and each application developer consumes those secrets in a way that is custom to their applications.  While there is a good deal of value to this level of flexibility, large development teams lose overall velocity dealing with each unique solution.  To combat this, we already see teams adopting internal patterns for how to achieve this application-to-service linkage.

The goal of this specification is to create a Kubernetes-wide specification for communicating service secrets to applications in an automated way.  It aims to create a mechanism that is widely applicable, but _without_ excluding other strategies for systems that it does not fit easily.  The benefit of a Kubernetes-wide specification is that all of the actors in an ecosystem can work towards a clearly defined abstraction at the edge of their expertise and depend on other parties to complete the chain.

* Application Developers expect secrets to be exposed in a consistent and predictable way
* Service Providers expect their secrets to be collected and exposed to users in a consistent and predictable way
* Platforms expect to retrieve secrets from Service Providers and expose them to Application Developers in a consistent and predictable way

The pattern of Service Binding has prior art in non-Kubernetes platforms.  Heroku pioneered this model with [Add-ons][h] and Cloud Foundry adopted similar ideas with their [Services][cf]. Other open source projects like the [Open Service Broker][osb] aim to help with this pattern on those non-Kubernetes platforms.  In the Kubernetes ecosystem, the CNCF Sandbox Cloud Native Buildpacks project has proposed a [buildpack-specific specification][cnb] exclusively addressing the application developer portion of this pattern.

[h]: https://devcenter.heroku.com/articles/add-ons
[cf]: https://docs.cloudfoundry.org/devguide/services/
[osb]: https://www.openservicebrokerapi.org
[cnb]: https://github.com/buildpacks/spec/blob/master/extensions/bindings.md

---
<!--ts-->
   * [Service Binding Specification for Kubernetes](#service-binding-specification-for-kubernetes)
      * [Notational Conventions](#notational-conventions)
      * [Terminology definition](#terminology-definition)
   * [Provisioned Service](#provisioned-service)
      * [Resource Type Schema](#resource-type-schema)
      * [Example Resource](#example-resource)
      * [Well-known Secret Entries](#well-known-secret-entries)
   * [Application Projection](#application-projection)
      * [Example Directory Structure](#example-directory-structure)
   * [Service Binding](#service-binding)
      * [Resource Type Schema](#resource-type-schema)
      * [Example Resource](#example-resource-1)
      * [Reconciler Implementation](#reconciler-implementation)
   * [Extensions](#extensions)
      * [Mapping Existing Values to New Values](#mapping-existing-values-to-new-values)
         * [Resource Type Schema](#resource-type-schema-1)
         * [Example Resource](#example-resource-2)
      * [Binding Values as Environment Variables](#binding-values-as-environment-variables)
         * [Resource Type Schema](#resource-type-schema-2)
         * [Example Resource](#example-resource-3)
      * [<kbd>EXPERIMENTAL</kbd> Synthetic data bindings](#experimental-synthetic-data-bindings)
      * [Subscription-based services](#subscription-based-services)

<!-- Added by: bhale, at: Mon Jun  1 20:46:50 PDT 2020 -->

<!--te-->

## Notational Conventions

The key words "MUST", "MUST NOT", "REQUIRED", "SHALL", "SHALL NOT", "SHOULD", "SHOULD NOT", "RECOMMENDED", "NOT RECOMMENDED", "MAY", and "OPTIONAL" are to be interpreted as described in [RFC 2119](http://tools.ietf.org/html/rfc2119).

The key words "unspecified", "undefined", and "implementation-defined" are to be interpreted as described in the [rationale for the C99 standard](http://www.open-std.org/jtc1/sc22/wg14/www/C99RationaleV5.10.pdf#page=18).

An implementation is not compliant if it fails to satisfy one or more of the MUST, MUST NOT, REQUIRED, SHALL, or SHALL NOT requirements for the protocols it implements.  An implementation is compliant if it satisfies all the MUST, MUST NOT, REQUIRED, SHALL, and SHALL NOT requirements for the protocols it implements.

## Terminology definition

<dl>
  <dt>Duck Type</dt>
  <dd>Any type that meets the contract defined in a specification, without being an instance of a specific concrete type.  For example, for specification that requires a given key on <code>status</code>, any resource that has that key on its <code>status</code> regardless of its <code>kind</code> would be considered a duck type of the specification.</dd>

  <dt>Service</dt>
  <dd>Any software that exposes functionality.  Examples include an application with REST endpoints, an event stream, an Application Performance Monitor, or a Hardware Security Module.</dd>

  <dt>Application</dt>
  <dd>Any process, running within a container.  Examples include a Spring Boot application, a NodeJS Express application, or a Ruby Rails application.  <b>Note:</b> This is different than an umbrella application as defined by the Kubernetes SIG, which refers to a set of micro-services.</dd>

  <dt>Service Binding</dt>
  <dd>The act of or representation of the action of providing information about a Service to an Application</dd>

  <dt>ConfigMap</dt>
  <dd>A Kubernetes <a href="https://kubernetes.io/docs/concepts/configuration/secret/">ConfigMap</a></dd>

  <dt>Secret</dt>
  <dd>A Kubernetes <a href="https://kubernetes.io/docs/concepts/configuration/configmap/">Secret</a></dd>
</dl>

# Provisioned Service

A Provisioned Service resource **MUST** define a `.status.binding.name` which is a `LocalObjectReference` to a `Secret`.  The `Secret` **MUST** be in the same namespace as the resource.  The `Secret` **MUST** contain a `kind` entry with a value that identifies the abstract classification of the binding.  It is **RECOMMENDED** that the `Secret` also contain a `provider` entry with a value that identifies the provider of the binding.  The `Secret` **MAY** contain any other entry.

## Resource Type Schema

```yaml
status:
  binding:
    name:  # string
```

## Example Resource

```yaml
...
status:
  ...
  binding:
    name: production-db-secret
```

## Well-known Secret Entries

Other than the required `kind` entry and the recommended `provider` entry, there are no other reserved `Secret` entries.  In the interests of consistency, if a `Secret` includes any of the following entry names, the entry value **MUST** meet the specified requirements:

| Name | Requirements
| ---- | ------------
| `host` | A DNS-resolvable host name or IP address
| `port` | A valid port number
| `uri` | A valid URI as defined by [RFC3986](https://tools.ietf.org/html/rfc3986)
| `username` | A string-based username credential
| `password` | A string-based password credential
| `certificates` | A collection of PEM-encoded X.509 certificates, representing a certificate chain used in mTLS client authentication
| `privateKey` | A PEM-encoded private key used in mTLS client authentication

`Secret` entries that do not meet these requirements **MUST** use different entry names.

# Application Projection

A Binding `Secret` **MUST** be volume mounted into a container at in `$SERVICE_BINDINGS_ROOT/<binding-name>` with directory names matching the name of the binding.  Binding names **MUST** match `[a-z0-9\-\.]{1,253}`.  The `$SERVICE_BINDINGS_ROOT` environment variable **MUST** be declared and can point to any valid file system location.

The `Secret` **MUST** contain a `kind` entry with a value that identifies the abstract classification of the binding.  It is **RECOMMENDED** that the `Secret` also contain a `provider` entry with a value that identifies the provider of the binding.  The `Secret` **MAY** contain any other entry.

The contents of a secret entry may be anything representable as bytes on the file system including, but not limited to, a literal string value (e.g. `db-password`), a language-specific binary (e.g. a Java `KeyStore` with a private key and X.509 certificate), or an indirect pointer to another system for value resolution (e.g. `vault://production-database/password`).

The collection of files within the directory **MAY** change between container launches.  The collection of files within the directory **MUST NOT** change during the lifetime of the container.

## Example Directory Structure

```plain
$SERVICE_BINDING_ROOT
├── account-database
│   ├── kind
│   ├── connection-count
│   ├── uri
│   ├── username
│   └── password
└── transaction-event-stream
│   ├── kind
│   ├── connection-count
│   ├── uri
│   ├── certificates
│   └── private-key
```

# Service Binding

A Service Binding describes the connection between a [Provisioned Service](#provisioned-service) and an [Application Projection](#application-projection).  It is codified as a concrete resource type.  Multiple Service Bindings can refer to the same service.  Multiple Service Bindings can refer to the same application.

A Service Binding resource **MUST** define a `.spec.application` which is an `ObjectReference` to a `PodSpec`-able resource.  A Service Binding resource **MUST** define a `.spec.service` which is an `ObjectReference` to a Provisioned Service-able resource.  A Service Binding resource **MAY** define a `.spec.name` which is the name of the service when projected into to the application.

A Service Binding resource **MUST** define a `.status.conditions` which is an array of `Condition` objects.  A `Condition` object **MUST** define `type`, `status`, and `lastTransitionTime` entries.  At least one condition containing a `type` of `Ready` must be defined.  The `status` of the `Ready` condition **MUST** have a value of `True`, `False`, or `Unknown`.  The `lastTranstionTime` **MUST** contain the last time that the condition transitioned from one status to another.  A Service Binding resource **MAY** define `reason` and `message` entries to describe the last `status` transition.

## Resource Type Schema

```yaml
apiVersion: service.binding/v1alpha1
kind: ServiceBinding
metadata:
  name:                 # string
spec:
  name:                 # string, optional, default: .metadata.name
  kind:                 # string, optional
  provider:             # string, optional

  application:          # PodSpec-able resource ObjectReference
    apiVersion:         # string
    kind:               # string
    name:               # string
    ...

  service:              # Provisioned Service-able resource ObjectReference
    apiVersion:         # string
    kind:               # string
    name:               # string
    ...

status:
  conditions:           # []Condition containing at least one entry for `Ready`
  - type:               # string
    status:             # string
    lastTransitionTime: # Time
    reason:             # string
    message:            # string
```

## Example Resource

```yaml
apiVersion: service.binding/v1alpha1
kind: ServiceBinding
metadata:
  name: online-banking-to-account-service
spec:
  name: account-service

  application:
    apiVersion: apps/v1
    kind:       Deployment
    name:       online-banking

  service:
    apiVersion: com.example/v1alpha1
    kind:       AccountService
    name:       prod-account-service

status:
  conditions:
  - type:   Ready
    status: True
```

## Reconciler Implementation

A Reconciler implementation for the `ServiceBinding` type is responsible for binding the Provisioned Service binding `Secret` into an Application.  The `Secret` referred to by `.status.binding.name` on the resource represented by `service` **MUST** be mounted as a volume on the resource represented by `application`.

If the `$SERVICE_BINDING_ROOT` environment variable has already been configured on the resource represented by `application`, the Provisioned Service binding `Secret` **MUST** be mounted relative to that location.  If the `$SERVICE_BINDING_ROOT` environment variable has not been configured on the resource represented by `application`, the `$SERVICE_BINDING_ROOT` environment variable **MUST** be set and the Provisioned Service binding `Secret` **MUST** be mounted relative to that location.

The `$SERVICE_BINDING_ROOT` environment variable **MUST NOT** be reset if it is already configured on the resource represented by `application`.

If a `.spec.name` is set, the directory name relative to `$SERVICE_BINDING_ROOT` **MUST** be its value.  If a `.spec.name` is not set, the directory name relative to `$SERVICE_BINDING_ROOT` **SHOULD** be the value of `.metadata.name`.

If a `.spec.kind` is set, the `kind` entry in the binding `Secret` **MUST** be set to its value overriding any existing value.  If a `.spec.provider` is set, the `provider` entry in the binding `Secret` **MUST** be set to its value overriding any existing value.

If the modification of the Application resource is completed successfully, the `Ready` condition status **MUST** be set to `True`.  If the modification of the Application resource is not completed sucessfully the `Ready` condition status **MUST NOT** be set to `True`.

# Extensions

Extensions are optional additions to the core specification as defined above.  Implementation and support of these specifications are not required in order for a platform to be considered compliant.  However, if the features addressed by these specifications are supported a platform **MUST** be in compliance with the specification that governs that feature.

## Mapping Existing Values to New Values

Many applications will not be able to consume the secrets exposed by Provisioned Services directly.  Teams creating Provisioned Services do not know how their services will be consumed, teams creating Applications will not know what services will be provided to them, different language families have different idioms for naming and style, and more.  Users must have a way of describing a mapping from existing values to customize the provided entries to ones that are usable directly by their applications.  This specification is described as an extension to the [Service Binding](#service-binding) specification and assumes full compatibility with it.

A Service Binding Resource **MAY** define a `.spec.mappings` which is an array of `Mapping` objects.  A `Mapping` object **MUST** define `name` and `value` entries.  The value of a `Mapping` **MAY** contain zero or more tokens beginning with `((`, ending with `))`, and encapsulating a binding `Secret` key name.  The value of this `Secret` entry **MUST** be substituted into the original `value` string, replacing the token.  Once all tokens have been substituted, the new `value` **MUST** be added to the `Secret` exposed to the resource represented by `application`.

### Resource Type Schema

```yaml
apiVersion: service.binding/v1alpha1
kind: ServiceBinding
metadata:
  name:         # string
spec:
  name:         # string, optional, default: .metadata.name

  application:  # PodSpec-able resource ObjectReference
    apiVersion: # string
    kind:       # string
    name:       # string
    ...

  service:      # Provisioned Service-able resource ObjectReference
    apiVersion: # string
    kind:       # string
    name:       # string
    ...

  mapping:      # []Mapping, optional
  - name:       # string
    value:      # string
  ...
```

### Example Resource

```yaml
apiVersion: service.binding/v1alpha1
kind: ServiceBinding
metadata:
  name: online-banking-to-account-service
spec:
  name: account-service

  application:
    apiVersion: apps/v1
    kind:       Deployment
    name:       online-banking

  service:
    apiVersion: com.example/v1alpha1
    kind:       AccountService
    name:       prod-account-service

  mapping:
  - name:  accountServiceUri
    value: https://((username)):((password))@((host)):((port))/((path))
```

## Binding Values as Environment Variables

Many applications, especially initially, will not be able to consume Service Bindings as defined by the Application Projection section directly since many of these applications assume that configuration will be exposed via environment variables.  Users must have a way of describing how they would like environment variables containing the values from bound secrets mapped into their applications.  This specification is described as an extension to the [Service Binding](#service-binding) specification and assumes full compatibility with it.

A Service Binding Resource **MAY** define a `.spec.env` which is an array of `EnvVar`.  The value of an entry in this array **MAY** contain zero or more tokens beginning with `((`, ending with `))`, and encapsulating a binding `Secret` key name.  The value of this `Secret` entry **MUST** be substituted into the original `value` string, replacing the token.  Once all tokens have been substituted, the new `value` **MUST** be configured as an environment variable on the resource represented by `application`.

### Resource Type Schema

```yaml
apiVersion: service.binding/v1alpha1
kind: ServiceBinding
metadata:
  name:         # string
spec:
  name:         # string, optional, default: .metadata.name

  application:  # PodSpec-able resource ObjectReference
    apiVersion: # string
    kind:       # string
    name:       # string
    ...

  service:      # Provisioned Service-able resource ObjectReference
    apiVersion: # string
    kind:       # string
    name:       # string
    ...

  env:          # []EnvVar, optional
  - name:       # string
    value:      # string
  ...
```

### Example Resource

```yaml
apiVersion: service.binding/v1alpha1
kind: ServiceBinding
metadata:
  name: online-banking-to-account-service
spec:
  name: account-service

  application:
    apiVersion: apps/v1
    kind:       Deployment
    name:       online-banking

  service:
    apiVersion: com.example/v1alpha1
    kind:       AccountService
    name:       prod-account-service

  env:
  - name:  ACCOUNT_SERVICE_HOST
    value: ((host))
  - name:  ACCOUNT_SERVICE_USERNAME
    value: ((username))
  - name:  ACCOUNT_SERVICE_PASSWORD
    value: ((password))
  - name:  ACCOUNT_SERVICE_URI
    value: ((accountServiceUri))
```



---




Binding is requested by the consuming application, or an entity on its behalf such as the [Runtime Component Operator](https://github.com/application-stacks/runtime-component-operator), via a custom resource that is applied in the same cluster where an implementation of this specification resides.

Since the reference implementation for most of this specification is the [Service Binding Operator](https://github.com/redhat-developer/service-binding-operator) we will be using the `ServiceBinding` CRD, which resides in [this folder](https://github.com/redhat-developer/service-binding-operator/tree/master/deploy/crds), as the entity that holds the binding request.

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

## <kbd>EXPERIMENTAL</kbd> Synthetic data bindings

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
          value: {{  event-cluster.status.url }} / ?username= {{ event-user.status.username }}
```

The service entry with apiVersion `servicebinding/v1alpha1` and kind `ComposedService` refers to a synthetic CR whose sole purpose is to compose bindings from other services.

## Subscription-based services

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


<!--
  ## Minimum requirements for being bindable

* include a sample `ServiceBinding` (see the [Request service binding](#request-service-binding) section below) in its documentation (e.g. GitHub repository, installation instructions, etc) which contains either:
  * a `dataMapping` element illustrating how each of its `status`, `spec` or `data` properties map to the corresponding [binding data](#service-binding-schema).
  * a `detectBindingResources: true` element which will automatically populate the resulting Secret from the `ServiceBinding` with information from any Route, Ingress, Service, ConfigMap or Secret resources that are owned by the backing service CR.

<kbd>EXPERIMENTAL</kbd>The service **MUST** also make itself discoverable by complying with one-of:
* In the case of an OLM-based Operator, add `Bindable` to the CSV's `metadata.annotations.categories`.
* In the case of a Helm chart service, add bindable to the Chart.yaml's keyword list.
* In all other cases, add the `servicebinding/bindable: "true"` annotation to your CRD or any CR (Secret, Service, etc).

## Pointer to binding data

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

Extra binding properties **SHOULD** also be defined by the bindable service, using one of the patterns defined in [Pointer to binding data](#pointer-to-binding-data).


-->
