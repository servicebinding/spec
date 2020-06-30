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

<!-- Using https://github.com/yzhang-gh/vscode-markdown to manage toc -->
- [Service Binding Specification for Kubernetes](#service-binding-specification-for-kubernetes)
  - [Notational Conventions](#notational-conventions)
  - [Terminology definition](#terminology-definition)
- [Provisioned Service](#provisioned-service)
  - [Resource Type Schema](#resource-type-schema)
  - [Example Resource](#example-resource)
  - [Well-known Secret Entries](#well-known-secret-entries)
  - [Example Secret](#example-secret)
- [Application Projection](#application-projection)
  - [Example Directory Structure](#example-directory-structure)
- [Service Binding](#service-binding)
  - [Resource Type Schema](#resource-type-schema-1)
  - [Minimal Example Resource](#minimal-example-resource)
  - [Label Selector Example Resource](#label-selector-example-resource)
  - [Mappings Example Resource](#mappings-example-resource)
  - [Environment Variables Example Resource](#environment-variables-example-resource)
  - [Reconciler Implementation](#reconciler-implementation)
- [Extensions](#extensions)
  - [Binding `Secret` Generation Strategies](#binding-secret-generation-strategies)
    - [OLM Operator Descriptors](#olm-operator-descriptors)
    - [Descriptor Examples](#descriptor-examples)
    - [Non-OLM Operator and Resource Annotations](#non-olm-operator-and-resource-annotations)
    - [Annotation Examples](#annotation-examples)
  - [Role-Based Access Control (RBAC)](#role-based-access-control-rbac)
    - [For Cluster Operators and CRD Authors](#for-cluster-operators-and-crd-authors)
      - [Example Resource](#example-resource-1)
    - [For Service Binding Implementors](#for-service-binding-implementors)
      - [Example Resource](#example-resource-2)

---

## Notational Conventions

The key words "MUST", "MUST NOT", "REQUIRED", "SHALL", "SHALL NOT", "SHOULD", "SHOULD NOT", "RECOMMENDED", "NOT RECOMMENDED", "MAY", and "OPTIONAL" are to be interpreted as described in [RFC 2119](http://tools.ietf.org/html/rfc2119).

The key words "unspecified", "undefined", and "implementation-defined" are to be interpreted as described in the [rationale for the C99 standard](http://www.open-std.org/jtc1/sc22/wg14/www/C99RationaleV5.10.pdf#page=18).

An implementation is not compliant if it fails to satisfy one or more of the MUST, MUST NOT, REQUIRED, SHALL, or SHALL NOT requirements for the protocols it implements.  An implementation is compliant if it satisfies all the MUST, MUST NOT, REQUIRED, SHALL, and SHALL NOT requirements for the protocols it implements.

## Terminology definition

<dl>
  <dt>Duck Type</dt>
  <dd>Any type that meets the contract defined in a specification, without being an instance of a specific concrete type.  For example, for specification that requires a given key on <code>status</code>, any resource that has that key on its <code>status</code> regardless of its <code>kind</code> would be considered a duck type of the specification.</dd>

  <dt>Service</dt>
  <dd>Any software that exposes functionality.  Examples include a database, a message broker, an application with REST endpoints, an event stream, an Application Performance Monitor, or a Hardware Security Module.</dd>

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

A Provisioned Service resource **MUST** define a `.status.binding.name` which is a `LocalObjectReference`-able to a `Secret`.  The `Secret` **MUST** be in the same namespace as the resource.  The `Secret` **SHOULD** contain a `type` entry with a value that identifies the abstract classification of the binding.  It is **RECOMMENDED** that the `Secret` also contain a `provider` entry with a value that identifies the provider of the binding.  The `Secret` **MAY** contain any other entry.

Extensions and implementations **MAY** define additional mechanisms to consume a Provisioned Service that does not conform to the duck type.

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

Other than the recommended `type` and `provider` entries, there are no other reserved `Secret` entries.  In the interests of consistency, if a `Secret` includes any of the following entry names, the entry value **MUST** meet the specified requirements:

| Name | Requirements
| ---- | ------------
| `host` | A DNS-resolvable host name or IP address
| `port` | A valid port number
| `uri` | A valid URI as defined by [RFC3986](https://tools.ietf.org/html/rfc3986)
| `username` | A string-based username credential
| `password` | A string-based password credential
| `certificates` | A collection of PEM-encoded X.509 certificates, representing a certificate chain used in mTLS client authentication
| `private-key` | A PEM-encoded private key used in mTLS client authentication

`Secret` entries that do not meet these requirements **MUST** use different entry names.

## Example Secret

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: production-db
stringData:
  type: mysql
  provider: bitnami
  host: localhost
  port: 3306
  username: root
  password: root
```

# Application Projection

A Binding `Secret` **MUST** be volume mounted into a container at `$SERVICE_BINDINGS_ROOT/<binding-name>` with directory names matching the name of the binding.  Binding names **MUST** match `[a-z0-9\-\.]{1,253}`.  The `$SERVICE_BINDINGS_ROOT` environment variable **MUST** be declared and can point to any valid file system location.

The `Secret` **MUST** contain a `type` entry with a value that identifies the abstract classification of the binding.  It is **RECOMMENDED** that the `Secret` also contain a `provider` entry with a value that identifies the provider of the binding.  The `Secret` **MAY** contain any other entry.

The name of a secret entry file name **SHOULD** match `[a-z0-9\-\.]{1,253}`.  The contents of a secret entry may be anything representable as bytes on the file system including, but not limited to, a literal string value (e.g. `db-password`), a language-specific binary (e.g. a Java `KeyStore` with a private key and X.509 certificate), or an indirect pointer to another system for value resolution (e.g. `vault://production-database/password`).

The collection of files within the directory **MAY** change between container launches.  The collection of files within the directory **SHOULD NOT** change during the lifetime of the container.

## Example Directory Structure

```plain
$SERVICE_BINDING_ROOT
├── account-database
│   ├── type
│   ├── provider
│   ├── uri
│   ├── username
│   └── password
└── transaction-event-stream
    ├── type
    ├── connection-count
    ├── uri
    ├── certificates
    └── private-key
```

# Service Binding

A Service Binding describes the connection between a [Provisioned Service](#provisioned-service) and an [Application Projection](#application-projection).  It is codified as a concrete resource type.  Multiple Service Bindings can refer to the same service.  Multiple Service Bindings can refer to the same application.

Restricting service binding to resources within the same namespace is strongly **RECOMMENDED**.  Cross-namespace service binding **SHOULD** be secured appropriately by the implementor to prevent attacks like privilege escalation and secret enumeration.

A Service Binding resource **MUST** define a `.spec.application` which is an `ObjectReference`-able declaration to a `PodSpec`-able resource.  A `ServiceBinding` **MAY** define the application reference by-name or by-[label selector][ls]. A name and selector **MUST NOT** be defined in the same reference.  A Service Binding resource **MUST** define a `.spec.service` which is an `ObjectReference`-able declaration to a Provisioned Service-able resource.

A Service Binding Resource **MAY** define a `.spec.mappings` which is an array of `Mapping` objects.  A `Mapping` object **MUST** define `name` and `value` entries.  The `value` of a `Mapping` **MAY** contain zero or more tokens beginning with `((`, ending with `))`, and encapsulating a binding `Secret` key name.  The value of this `Secret` entry **MUST** be substituted into the original `value` string, replacing the token.  Once all tokens have been substituted, the new `value` **MUST** be added to the `Secret` exposed to the resource represented by `application`.

A Service Binding Resource **MAY** define a `.spec.env` which is an array of `EnvVar`.  An `EnvVar` object **MUST** define `name` and `key` entries.  The `key` of an `EnvVar` **MUST** refer to a binding `Secret` key name including any key defined by a `Mapping`.  The value of this `Secret` entry **MUST** be configured as an environment variable on the resource represented by `application`.

A Service Binding resource **MUST** define a `.status.conditions` which is an array of `Condition` objects.  A `Condition` object **MUST** define `type`, `status`, and `lastTransitionTime` entries.  At least one condition containing a `type` of `Ready` must be defined.  The `status` of the `Ready` condition **MUST** have a value of `True`, `False`, or `Unknown`.  The `lastTransitionTime` **MUST** contain the last time that the condition transitioned from one status to another.  A Service Binding resource **MAY** define `reason` and `message` entries to describe the last `status` transition.  As label selectors are inherently queries that return zero-to-many resources, it is **RECOMMENDED** that `ServiceBinding` authors use a combination of labels that yield a single resource, but implementors **MUST** handle each matching resource as if it was specified by name in a distinct `ServiceBinding` resource. Partial failures **MUST** be aggregated and reported on the binding status's `Ready` condition. A Service Binding resource **MAY** reflect the secret projected into the application as `.status.binding.name`.

[ls]: https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/#label-selectors

## Resource Type Schema

```yaml
apiVersion: service.binding/v1alpha1
kind: ServiceBinding
metadata:
  name:                 # string
spec:
  name:                 # string, optional, default: .metadata.name
  type:                 # string, optional
  provider:             # string, optional

  application:          # PodSpec-able resource ObjectReference-able
    apiVersion:         # string
    kind:               # string
    name:               # string, mutually exclusive with selector
    selector:           # metav1.LabelSelector, mutually exclusive with name
    containers:         # []intstr.IntOrString, optional
    ...

  service:              # Provisioned Service-able resource ObjectReference-able
    apiVersion:         # string
    kind:               # string
    name:               # string
    ...

  mappings:             # []Mapping, optional
  - name:               # string
    value:              # string
  ...

  env:                  # []EnvVar, optional
  - name:               # string
    key:                # string
  ...

status:
  binding:              # LocalObjectReference, optional
    name:               # string
  conditions:           # []Condition containing at least one entry for `Ready`
  - type:               # string
    status:             # string
    lastTransitionTime: # Time
    reason:             # string
    message:            # string
```

## Minimal Example Resource

```yaml
apiVersion: service.binding/v1alpha1
kind: ServiceBinding
metadata:
  name: account-service
spec:
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

## Label Selector Example Resource

```yaml
apiVersion: service.binding/v1alpha1
kind: ServiceBinding
metadata:
  name: online-banking-frontend-to-account-service
spec:
  name: account-service

  application:
    apiVersion: apps/v1
    kind:       Deployment
    selector:
      matchLabels:
        app.kubernetes.io/part-of: online-banking
        app.kubernetes.io/component: frontend

  service:
    apiVersion: com.example/v1alpha1
    kind:       AccountService
    name:       prod-account-service

status:
  conditions:
  - type:   Ready
    status: True
```

## Mappings Example Resource

```yaml
apiVersion: service.binding/v1alpha1
kind: ServiceBinding
metadata:
  name: account-service
spec:
  application:
    apiVersion: apps/v1
    kind:       Deployment
    name:       online-banking

  service:
    apiVersion: com.example/v1alpha1
    kind:       AccountService
    name:       prod-account-service

  mappings:
  - name:  accountServiceUri
    value: https://((username)):((password))@((host)):((port))/((path))

status:
  binding:
    name: prod-account-service-projection
  conditions:
  - type:   Ready
    status: True
```

## Environment Variables Example Resource

```yaml
apiVersion: service.binding/v1alpha1
kind: ServiceBinding
metadata:
  name: account-service
spec:
  application:
    apiVersion: apps/v1
    kind:       Deployment
    name:       online-banking

  service:
    apiVersion: com.example/v1alpha1
    kind:       AccountService
    name:       prod-account-service

  mappings:
  - name:  accountServiceUri
    value: https://((username)):((password))@((host)):((port))/((path))

  env:
  - name: ACCOUNT_SERVICE_HOST
    key:  host
  - name: ACCOUNT_SERVICE_USERNAME
    key:  username
  - name: ACCOUNT_SERVICE_PASSWORD
    key:  password
  - name: ACCOUNT_SERVICE_URI
    key:  accountServiceUri

status:
  conditions:
  - type:   Ready
    status: True
```

## Reconciler Implementation

A Reconciler implementation for the `ServiceBinding` type is responsible for binding the Provisioned Service binding `Secret` into an Application.  The `Secret` referred to by `.status.binding.name` on the resource represented by `service` **MUST** be mounted as a volume on the resource represented by `application`.  If the `application` resource is managed by another Reconciler, a `ServiceBinding` Implementations **SHOULD** ensure that the `Secret` volume mount configuration remains after the other Reconciler completes.

If a `.spec.name` is set, the directory name of the volume mount **MUST** be its value.  If a `.spec.name` is not set, the directory name of the volume mount **SHOULD** be the value of `.metadata.name`.

If the `$SERVICE_BINDING_ROOT` environment variable has already been configured on the resource represented by `application`, the Provisioned Service binding `Secret` **MUST** be mounted relative to that location.  If the `$SERVICE_BINDING_ROOT` environment variable has not been configured on the resource represented by `application`, the `$SERVICE_BINDING_ROOT` environment variable **MUST** be set and the Provisioned Service binding `Secret` **MUST** be mounted relative to that location.  A **RECOMMENDED** value to use is `/bindings`.

The `$SERVICE_BINDING_ROOT` environment variable **MUST NOT** be reset if it is already configured on the resource represented by `application`.

If a `.spec.type` is set, the `type` entry in the binding `Secret` **MUST** be set to its value overriding any existing value.  If a `.spec.provider` is set, the `provider` entry in the binding `Secret` **MUST** be set to its value overriding any existing value.

If the modification of the Application resource is completed successfully, the `Ready` condition status **MUST** be set to `True`.  If the modification of the Application resource is not completed successfully the `Ready` condition status **MUST NOT** be set to `True`.

# Extensions

Extensions are optional additions to the core specification as defined above.  Implementation and support of these specifications are not required in order for a platform to be considered compliant.  However, if the features addressed by these specifications are supported a platform **MUST** be in compliance with the specification that governs that feature.

## Binding `Secret` Generation Strategies

Many services, especially initially, will not be Provisioned Service-compliant.  These services will expose the appropriate binding `Secret` information, but not in the way that the specification or applications expect.  Users should have a way of describing a mapping from existing data associated with arbitrary resources and CRDs to a representation of a binding `Secret`.

To handle the majority of existing resources and CRDs, `Secret` generation needs to support the following behaviors:

1.  Extract a string from a resource
1.  Extract an entire `ConfigMap`/`Secret` refrenced from a resource
1.  Extract a specific entry in a `ConfigMap`/`Secret` referenced from a resource
1.  Extract entries from a collection of objects, mapping keys and values from entries in a `ConfigMap`/`Secret` referenced from a resource
1.  Map each value to a specific key

While the syntax of the generation strategies are specific to the system they are annotating, they are based on a common data model.

| Model | Description
| ----- | -----------
| `path` | A template represention of the path to an element in a Kubernetes resource.  The value of `path` is specified as [JSONPath](https://kubernetes.io/docs/reference/kubectl/jsonpath/).  Required.
| `objectType` | Specifies the type of the object selected by the `path`.  One of `ConfigMap`, `Secret`, or `string` (default).
| `elementType` | Specifies the type of object in an array selected by the `path`.  One of `sliceOfMaps`, `sliceOfStrings`, `string` (default).
| `sourceKey` | Specifies a particular key to select if a `ConfigMap` or `Secret` is selected by the `path`.  Specifies a value to use for the key for an entry in a binding `Secret` when `elementType` is `sliceOfMaps`.
| `sourceValue` | Specifies a particular value to use for the value for an entry in a binding `Secret` when `elementType` is `sliceOfMaps`


### OLM Operator Descriptors

OLM Operators are configured by setting the `specDescriptor` and `statusDescriptor` entries in the [ClusterServiceVersion](https://docs.openshift.com/container-platform/4.4/operators/operator_sdk/osdk-generating-csvs.html) with mapping descriptors.

### Descriptor Examples

The following examples refer to this resource definition.

```yaml
apiVersion: apps.kube.io/v1beta1
kind: Database
metadata:
  name: my-cluster
spec:
  ...

status:
  bootstrap:
  - type: plain
    url: myhost2.example.com
    name: hostGroup1
  - type: tls
    url: myhost1.example.com:9092,myhost2.example.com:9092
    name: hostGroup2
  data:
    dbConfiguration: database-config     # ConfigMap
    dbCredentials: database-cred-Secret  # Secret
    url: db.stage.ibm.com
```

1.  Mount an entire `Secret` as the binding `Secret`
    ```yaml
    - path: data.dbCredentials
      x-descriptors:
      - urn:alm:descriptor:io.kubernetes:Secret
      - service.binding
    ```
1.  Mount an entire `ConfigMap` as the binding `Secret`
    ```yanl
    - path: data.dbConfiguration
      x-descriptors:
      - urn:alm:descriptor:io.kubernetes:ConfigMap
      - service.binding
    ```
1.  Mount an entry from a `ConfigMap` into the binding `Secret`
    ```yaml
    - path: data.dbConfiguration
      x-descriptors:
      - urn:alm:descriptor:io.kubernetes:ConfigMap
      - service.binding:certificate:sourceKey=certificate
    ```
1.  Mount an entry from a `ConfigMap` into the binding `Secret` with a different key
    ```yaml
    - path: data.dbConfiguration
      x-descriptors:
      - urn:alm:descriptor:io.kubernetes:ConfigMap
      - servicebinding:timeout:sourceKey=db_timeout
    ```
1.  Mount a resource definition value into the binding `Secret`
    ```yaml
    - path: data.uri
      x-descriptors:
      - service.binding:uri
    ```
1.  Mount a resource definition value into the binding `Secret` with a different key
    ```yaml
    - path: data.connectionURL
      x-descriptors:
      - service.binding:uri
    ```
1.  Mount the entries of a collection into the binding `Secret` selecting the key and value from each entry
    ```yaml
    - path: bootstrap
      x-descriptors:
      - service.binding:endpoints:elementType=sliceOfMaps:sourceKey=type:sourceValue=url
    ```

### Non-OLM Operator and Resource Annotations

Non-OLM Operators are configured by adding annotations to the Operator's CRD with mapping configuration.  All Kubernetes resources are configured by adding annotations to the resource.

### Annotation Examples

The following examples refer to this resource definition.

```yaml
apiVersion: apps.kube.io/v1beta1
kind: Database
metadata:
  name: my-cluster
spec:
  ...

status:
  bootstrap:
  - type: plain
    url: myhost2.example.com
    name: hostGroup1
  - type: tls
    url: myhost1.example.com:9092,myhost2.example.com:9092
    name: hostGroup2
  data:
    dbConfiguration: database-config     # ConfigMap
    dbCredentials: database-cred-Secret  # Secret
    url: db.stage.ibm.com
```

1.  Mount an entire `Secret` as the binding `Secret`
    ```plain
    “service.binding":
      ”path={.status.data.dbCredentials},objectType=Secret”
    ```
1.  Mount an entire `ConfigMap` as the binding `Secret`
    ```plain
    service.binding”:
      "path={.status.data.dbConfiguration},objectType=ConfigMap”
    ```
1.  Mount an entry from a `ConfigMap` into the binding `Secret`
    ```plain
    “service.binding/certificate”:
      "path={.status.data.dbConfiguration},objectType=ConfigMap,sourceKey=certificate"
    ```
1.  Mount an entry from a `ConfigMap` into the binding `Secret` with a different key
    ```plain
    “service.binding/timeout”:
      “path={.status.data.dbConfiguration},objectType=ConfigMap,sourceKey=db_timeout”
    ```
1.  Mount a resource definition value into the binding `Secret`
    ```plain
    “service.binding/uri”:
      "path={.status.data.url}"
    ```
1.  Mount a resource definition value into the binding `Secret` with a different key
    ```plain
    “service.binding/uri":
      "path={.status.data.connectionURL}”
    ```
1.  Mount the entries of a collection into the binding `Secret` selecting the key and value from each entry
    ```plain
    “service.binding/endpoints”:
      "path={.status.bootstrap},elementType=sliceOfMaps,sourceKey=type,sourceValue=url"
    ```

## Role-Based Access Control (RBAC)

Kubernetes clusters often utilize [Role-based access control (RBAC)][rbac] to authorize subjects to perform specific actions on resources. When operating in a cluster with RBAC enabled, the service binding reconciler needs permission to read resources that provisioned a service and write resources that services are projected into. This extension defines a means for third-party CRD authors and cluster operators to expose resources to the service binding reconciler. Cluster operators **MAY** impose additional access controls beyond RBAC.

[rbac]: https://kubernetes.io/docs/reference/access-authn-authz/rbac/

### For Cluster Operators and CRD Authors

Cluster operators and CRD authors **MAY** opt-in resources to service binding by defining a `ClusterRole` with a label matching `service.binding/controller=true`. For Provisioned Service-able resources the `get`, `list`, and `watch` verbs **MUST** be granted. For PodSpec-able resources the `get`, `list`, `watch`, `update`, and `patch` verbs **MUST** be granted.

#### Example Resource

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: awesome-service-bindings
  labels:
    service.binding/controller: "true" # matches the aggregation rule selector
rules:
# for Provisioned Service-able resources only
- apiGroups:
  - awesome.example.com
  resources:
  - awesomeservices
  verbs:
  - get
  - list
  - watch
# for PodSpec-able resources (also compatible with Provisioned Service-able resources)
- apiGroups:
  - awesome.example.com
  resources:
  - awesomeapplications
  verbs:
  - get
  - list
  - watch
  - update
  - patch
```

### For Service Binding Implementors

Service binding reconciler implementations **MUST** define an [aggregated `ClusterRole`][acr] with a label selector matching the label `service.binding/controller=true`. This `ClusterRole` **MUST** be bound (`RoleBinding` for a single namespace or `ClusterRoleBinding` if cluster-wide) to the subject the service binding reconciler runs as, typically a `ServiceAccount`.

[acr]: https://kubernetes.io/docs/reference/access-authn-authz/rbac/#aggregated-clusterroles

#### Example Resource

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: ...
aggregationRule:
  clusterRoleSelectors:
  - matchLabels:
      service.binding/controller: "true"
rules: [] # The control plane automatically fills in the rules
```
