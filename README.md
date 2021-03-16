# Service Binding Specification for Kubernetes

Today in Kubernetes, the exposure of secrets for connecting applications to external services such as REST APIs, databases, event buses, and many more is manual and bespoke.  Each service provider suggests a different way to access their secrets, and each application developer consumes those secrets in a custom way to their applications.  While there is a good deal of value to this flexibility level, large development teams lose overall velocity dealing with each unique solution.  To combat this, we already see teams adopting internal patterns for how to achieve this application-to-service linkage.

This specification aims to create a Kubernetes-wide specification for communicating service secrets to applications in an automated way.  It aims to create a widely applicable mechanism but _without_ excluding other strategies for systems that it does not fit easily.  The benefit of Kubernetes-wide specification is that all of the actors in an ecosystem can work towards a clearly defined abstraction at the edge of their expertise and depend on other parties to complete the chain.

* Application Developers expect their secrets to be exposed consistently and predictably.
* Service Providers expect their secrets to be collected and exposed to users consistently and predictably.
* Platforms expect to retrieve secrets from Service Providers and expose them to Application Developers consistently and predictably.

The pattern of Service Binding has prior art in non-Kubernetes platforms.  Heroku pioneered this model with [Add-ons][h], and Cloud Foundry adopted similar ideas with their [Services][cf].  Other open source projects like the [Open Service Broker][osb] aim to help with this pattern on those non-Kubernetes platforms.  In the Kubernetes ecosystem, the CNCF Sandbox Cloud Native Buildpacks project has proposed a [buildpack-specific specification][cnb] exclusively addressing the application developer portion of this pattern.

[h]: https://devcenter.heroku.com/articles/add-ons
[cf]: https://docs.cloudfoundry.org/devguide/services/
[osb]: https://www.openservicebrokerapi.org
[cnb]: https://github.com/buildpacks/spec/blob/master/extensions/bindings.md

<!-- omit in toc -->
### Community

The Service Binding Specification for Kubernetes project is a community lead effort.
A bi-weekly [working group call][working-group] is open to the public.
Discussions occur here on GitHub and on the [#bindings-discuss channel in the Kubernetes Slack][slack].

If you catch an error in the specification’s text, or if you write an
implementation, please let us know by opening an issue or pull request at our
[GitHub repository][repo].

Behavior within the project is governed by the [Contributor Covenant Code of Conduct][code-of-conduct].

[working-group]: https://docs.google.com/document/d/1rR0qLpsjU38nRXxeich7F5QUy73RHJ90hnZiFIQ-JJ8/edit#heading=h.ar8ibc31ux6f
[slack]: https://kubernetes.slack.com/archives/C012F2GPMTQ
[repo]: https://github.com/k8s-service-bindings/spec
[code-of-conduct]: ./CODE_OF_CONDUCT.md

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
    - [Ready Condition Status](#ready-condition-status)
- [Direct Secret Reference](#direct-secret-reference)
  - [Direct Secret Reference Example Resource](#direct-secret-reference-example-resource)
- [Application Resource Mapping](#application-resource-mapping)
  - [Resource Type Schema](#resource-type-schema-2)
  - [Container-based Example Resource](#container-based-example-resource)
  - [Element-based Example Resource](#element-based-example-resource)
  - [PodSpec-able (Default) Example Resource](#podspec-able-default-example-resource)
  - [Reconciler Implementation](#reconciler-implementation-1)
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

The key words "MUST", "MUST NOT", "REQUIRED", "SHALL", "SHALL NOT", "SHOULD", "SHOULD NOT", "RECOMMENDED", "NOT RECOMMENDED", "MAY", and "OPTIONAL" in this document are to be interpreted as described in [BCP 14](https://tools.ietf.org/html/bcp14) [[RFC2119](https://tools.ietf.org/html/rfc2119)] [[RFC8174](https://tools.ietf.org/html/rfc8174)] when, and only when, they appear in all capitals, as shown here.

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
  <dd>A Kubernetes <a href="https://kubernetes.io/docs/concepts/configuration/configmap/">ConfigMap</a></dd>

  <dt>Secret</dt>
  <dd>A Kubernetes <a href="https://kubernetes.io/docs/concepts/configuration/secret/">Secret</a></dd>
</dl>

# Provisioned Service

A Provisioned Service resource **MUST** define a `.status.binding` which is a `LocalObjectReference`-able (containing a single field `name`) to a `Secret`.  The `Secret` **MUST** be in the same namespace as the resource.  The `Secret` data **SHOULD** contain a `type` entry with a value that identifies the abstract classification of the binding.  The `Secret` type (`.type` verses `.data.type`) **SHOULD** reflect this value as `service.binding/{type}`, replacing `{type}` with the `Secret` data type.  It is **RECOMMENDED** that the `Secret` data also contain a `provider` entry with a value that identifies the provider of the binding.  The `Secret` data **MAY** contain any other entry.  To facilitate discoverability, it is **RECOMMENDED** that a `CustomResourceDefinition` exposing a Provisioned Service add `service.binding/provisioned-service: "true"` as a label.

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
type: service.binding/mysql
stringData:
  type: mysql
  provider: bitnami
  host: localhost
  port: 3306
  username: root
  password: root
```

# Application Projection

A Binding `Secret` **MUST** be volume mounted into a container at `$SERVICE_BINDING_ROOT/<binding-name>` with directory names matching the name of the binding.  Binding names **MUST** match `[a-z0-9\-\.]{1,253}`.  The `$SERVICE_BINDING_ROOT` environment variable **MUST** be declared and can point to any valid file system location.

The `Secret` data **MUST** contain a `type` entry with a value that identifies the abstract classification of the binding.  The `Secret` type (`.type` verses `.data.type`) **MUST** reflect this value as `service.binding/{type}`, replacing `{type}` with the `Secret` data type.  It is **RECOMMENDED** that the `Secret` data also contain a `provider` entry with a value that identifies the provider of the binding.  The `Secret` data **MAY** contain any other entry.

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

A Service Binding describes the connection between a [Provisioned Service](#provisioned-service) and an [Application Projection](#application-projection).  It **MUST** be codified as a concrete resource type with API version `service.binding/v1beta1` and kind `ServiceBinding`.  Multiple Service Bindings can refer to the same service.  Multiple Service Bindings can refer to the same application.  For portability, the schema **MUST** comply to the exemplar CRD found [here][sb-crd].

Restricting service binding to resources within the same namespace is strongly **RECOMMENDED**.  Implementations that choose to support cross-namespace service binding **SHOULD** provide a security model that prevents attacks like privilege escalation and secret enumeration, as well as a deterministic way to declare target namespaces.

A Service Binding resource **MUST** define a `.spec.application` which is an `ObjectReference`-like declaration.  A `ServiceBinding` **MAY** define the application reference by-name or by-[label selector][ls]. A name and selector **MUST NOT** be defined in the same reference.  A Service Binding resource **MUST** define a `.spec.service` which is an `ObjectReference`-like declaration to a Provisioned Service-able resource.  Extensions and implementations **MAY** allow additional kinds of applications and services to be referenced.

The Service Binding resource **MAY** define `.spec.application.containers`, as a list of integers or strings, to limit which containers in the application are bound.  Binding to a container is opt-in, unless `.spec.application.containers` is undefined then all containers **MUST** be bound.  For each item in the containers list:
- if the value is an integer (`${containerInteger}`), the container matching by index (`.spec.template.spec.containers[${containerInteger}]`) **MUST** be bound. Init containers **MUST NOT** be bound
- if the value is a string (`${containerString}`), a container or init container matching by name (`.spec.template.spec.containers[?(@.name=='${containerString}')]` or `.spec.template.spec.initContainers[?(@.name=='${containerString}')]`) **MUST** be bound
- values that do not match a container or init container **SHOULD** be ignored

A Service Binding Resource **MAY** define a `.spec.mappings` which is an array of `Mapping` objects.  A `Mapping` object **MUST** define `name` and `value` entries.  The `value` of a `Mapping` **MUST** be handled as a [Go Template][gt] exposing binding `Secret` keys for substitution. The executed output of the template **MUST** be added to the `Secret` exposed to the resource represented by `application` as the key specified by the `name` of the `Mapping`.  If the `name` of a `Mapping` matches that of a Provisioned Service `Secret` key, the value from `Mapping` **MUST** be used for binding.

A Service Binding Resource **MAY** define a `.spec.env` which is an array of `EnvMapping`.  An `EnvMapping` object **MUST** define `name` and `key` entries.  The `key` of an `EnvMapping` **MUST** refer to a binding `Secret` key name including any key defined by a `Mapping`.  The value of this `Secret` entry **MUST** be configured as an environment variable on the resource represented by `application`.

A Service Binding resource **MUST** define `.status.conditions` which is an array of `Condition` objects as defined in [meta/v1 Condition][mv1c].  At least one condition containing a `type` of `Ready` **MUST** be defined.  The `Ready` condition **SHOULD** contain appropriate values defined by the implementation.  As label selectors are inherently queries that return zero-to-many resources, it is **RECOMMENDED** that `ServiceBinding` authors use a combination of labels that yield a single resource, but implementors **MUST** handle each matching resource as if it was specified by name in a distinct `ServiceBinding` resource. Partial failures **MUST** be aggregated and reported on the binding status's `Ready` condition. A Service Binding resource **SHOULD** reflect the secret projected into the application as `.status.binding.name`.

[sb-crd]: service.binding_servicebindings.yaml
[ls]: https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/#label-selectors
[gt]: https://golang.org/pkg/text/template/#pkg-overview
[mv1c]: https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.20/#condition-v1-meta

## Resource Type Schema

```yaml
apiVersion: service.binding/v1beta1
kind: ServiceBinding
metadata:
  name:                 # string
  generation:           # int64, defined by the Kubernetes control plane
  ...
spec:
  name:                 # string, optional, default: .metadata.name
  type:                 # string, optional
  provider:             # string, optional

  application:          # ObjectReference-like
    apiVersion:         # string
    kind:               # string
    name:               # string, mutually exclusive with selector
    selector:           # metav1.LabelSelector, mutually exclusive with name
    containers:         # []intstr.IntOrString, optional

  service:              # Provisioned Service resource ObjectReference-like
    apiVersion:         # string
    kind:               # string
    name:               # string

  mappings:             # []Mapping, optional
  - name:               # string
    value:              # string

  env:                  # []EnvMapping, optional
  - name:               # string
    key:                # string

status:
  binding:              # LocalObjectReference, optional
    name:               # string
  conditions:           # []metav1.Condition containing at least one entry for `Ready`
  observedGeneration:   # int64
```

## Minimal Example Resource

```yaml
apiVersion: service.binding/v1beta1
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
    status: 'True'
    reason: 'Projected'
    message: ''
    lastTransitionTime: '2021-01-20T17:00:00Z'
```

## Label Selector Example Resource

```yaml
apiVersion: service.binding/v1beta1
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
    status: 'True'
    reason: 'Projected'
    message: ''
    lastTransitionTime: '2021-01-20T17:00:00Z'
```

## Mappings Example Resource

```yaml
apiVersion: service.binding/v1beta1
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
    value: https://{{ urlquery .username }}:{{ urlquery .password }}@{{ .host }}:{{ .port }}/{{ .path }}

status:
  binding:
    name: prod-account-service-projection
  conditions:
  - type:   Ready
    status: 'True'
    reason: 'Projected'
    message: ''
    lastTransitionTime: '2021-01-20T17:00:00Z'
```

## Environment Variables Example Resource

```yaml
apiVersion: service.binding/v1beta1
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
    value: https://{{ urlquery .username }}:{{ urlquery .password }}@{{ .host }}:{{ .port }}/{{ .path }}

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
  binding:
    name: prod-account-service-projection
  conditions:
  - type:   Ready
    status: 'True'
    reason: 'Projected'
    message: ''
    lastTransitionTime: '2021-01-20T17:00:00Z'
```

## Reconciler Implementation

A Reconciler implementation for the `ServiceBinding` type is responsible for binding the Provisioned Service binding `Secret` into an Application.  The `Secret` referred to by `.status.binding` on the resource represented by `service` **MUST** be mounted as a volume on the resource represented by `application`.

If a `.spec.name` is set, the directory name of the volume mount **MUST** be its value.  If a `.spec.name` is not set, the directory name of the volume mount **SHOULD** be the value of `.metadata.name`.

If the `$SERVICE_BINDING_ROOT` environment variable has already been configured on the resource represented by `application`, the Provisioned Service binding `Secret` **MUST** be mounted relative to that location.  If the `$SERVICE_BINDING_ROOT` environment variable has not been configured on the resource represented by `application`, the `$SERVICE_BINDING_ROOT` environment variable **MUST** be set and the Provisioned Service binding `Secret` **MUST** be mounted relative to that location.  A **RECOMMENDED** value to use is `/bindings`.

The `$SERVICE_BINDING_ROOT` environment variable **MUST NOT** be reset if it is already configured on the resource represented by `application`.

If a `.spec.type` is set, the `type` entry in the binding `Secret` **MUST** be set to its value overriding any existing value.  If a `.spec.provider` is set, the `provider` entry in the binding `Secret` **MUST** be set to its value overriding any existing value.

### Ready Condition Status

If the modification of the Application resource is completed successfully, the `Ready` condition status **MUST** be set to `True`.  If the modification of the Application resource is not completed successfully the `Ready` condition status **MUST NOT** be set to `True`.

# Direct Secret Reference

There are scenarios where an appropriate resource conforming to the Provisioned Service duck-type does not exist, but there is a `Secret` available for binding.  This feature allows a `ServiceBinding` resource to directly reference a `Secret`.

When the `.spec.service.kind` attribute is `Secret` and `.spec.service.apiVersion` is `v1`, the `.spec.service.name` attribute **MUST** be treated as `.status.binding.name` for a Provisioned Service.

## Direct Secret Reference Example Resource

```yaml
apiVersion: service.binding/v1beta1
kind: ServiceBinding
metadata:
  name: account-service

spec:
  application:
    apiVersion: apps/v1
    kind:       Deployment
    name:       online-banking

  service:
    apiVersion: v1
    kind:       Secret
    name:       prod-account-service-secret

status:
  binding:
    name: prod-account-service-reference
  conditions:
  - type:   Ready
    status: 'True'
    reason: 'Projected'
    message: ''
    lastTransitionTime: '2021-01-20T17:00:00Z'
```

# Application Resource Mapping

An Application Resource Mapping describes how to apply [Service Binding](#service-binding) transformations to an [Application Projection](#application-projection).  It **MUST** be codified as a concrete resource type with API version `service.binding/v1beta1` and kind `ClusterApplicationResourceMapping`.  For portability, the schema **MUST** comply to the exemplar CRD found [here][carm-crd].

An Application Resource Mapping **MUST** define its name using [CRD syntax][crd-syntax] (`<plural>.<group>`) for the resource that it defines a mapping for.  An Application Resource Mapping **MUST** define a `.spec.versions` which is an array of `Version` objects.  A `Version` object must define a `version` entry that represents a version of the mapped resource.  The `version` entry **MAY** contain a `*` wildcard which indicates that this mapping should be used for any version that does not have a mapping explicitly defined for it.  A `Version` object **MAY** define `.containers`, as an array of strings containing [JSONPath][jsonpath], that describes the location of [`[]Container`][container] arrays in the target resource.  A `Version` object **MAY** define `.envs`, as an array of strings containing [JSONPath][jsonpath], that describes the location of [`[]EnvVar`][envvar] arrays in the target resource.  A `Version` object **MAY** define `.volumeMounts`, as an array of strings containing [JSONPath][jsonpath], that describes the location of [`[]VolumeMount`][volumemount] arrays in the target resource.  A `Version` object **MUST** define `.volumes`, as a string containing [JSONPath][jsonpath], that describes the location of [`[]Volume`][volume] arrays in the target resource.

If an Application Resource Mapping defines `containers`, it **MUST NOT** define `.envs` and `.volumeMounts`.  If an Application resources does not define `containers`, it **MUST** define `.envs` and `.volumeMounts`.

[carm-crd]: service.binding_clusterapplicationresourcemappings.yaml
[container]: https://kubernetes.io/docs/reference/kubernetes-api/workloads-resources/container/
[crd-syntax]: https://kubernetes.io/docs/tasks/extend-kubernetes/custom-resources/custom-resource-definitions/#create-a-customresourcedefinition
[envvar]: https://kubernetes.io/docs/reference/kubernetes-api/workloads-resources/container/#environment-variables
[jsonpath]: https://kubernetes.io/docs/reference/kubectl/jsonpath/
[volume]: https://kubernetes.io/docs/reference/kubernetes-api/config-and-storage-resources/volume
[volumemount]: https://kubernetes.io/docs/reference/kubernetes-api/workloads-resources/container/#volumes

## Resource Type Schema

```yaml
apiVersion: service.binding/v1beta1
kind: ClusterApplicationResourceMapping
metadata:
  name:                 # string
  generation:           # int64, defined by the Kubernetes control plane
  ...
spec:
  versions:             # []Version
  - version:            # string
    containers:         # []string, optional
    envs:               # []string, optional
    volumeMounts:       # []string, optional
    volumes:            # string
```

## Container-based Example Resource

```yaml
apiVersion: service.binding/v1beta1
kind: ClusterApplicationResourceMapping
metadata:
 name:  cronjobs.batch
spec:
  versions:
  - version: "*"
    containers:
    - .spec.jobTemplate.spec.template.spec.containers
    - .spec.jobTemplate.spec.template.spec.initContainers
    volumes: .spec.jobTemplate.spec.template.spec.volumes
```

## Element-based Example Resource

```yaml
apiVersion: service.binding/v1beta1
kind: ClusterApplicationResourceMapping
metadata:
 name:  cronjobs.batch
spec:
  versions:
  - version: "*"
    envs:
    - .spec.jobTemplate.spec.template.spec.containers[*].env
    - .spec.jobTemplate.spec.template.spec.initContainers[*].env
    volumeMounts:
    - .spec.jobTemplate.spec.template.spec.containers[*].volumeMounts
    - .spec.jobTemplate.spec.template.spec.initContainers[*].volumeMounts
    volumes: .spec.jobTemplate.spec.template.spec.volumes
```

## PodSpec-able (Default) Example Resource

```yaml
apiVersion: service.binding/v1beta1
kind: ClusterApplicationResourceMapping
metadata:
  name: deployments.apps
spec:
  versions:
  - version: "*"
    containers:
    - .spec.template.spec.containers
    - .spec.template.spec.initContainers
    volumes: .spec.template.spec.volumes
```

## Reconciler Implementation

A reconciler implementation **MUST** support mapping to PodSpec-able resources without defining a Application Resource Mapping for those types.  If no Application Resource Mapping exists for the `ServiceBinding` application resource type and the application resource is not PodSpec-able, the reconciliation **MUST** fail.

If a `ClusterApplicationResourceMapping` defines `containers`, the reconciler **MUST** first resolve a set of candidate locations in the application resource addressed by the `ServiceBinding` using the `Container` type (`.envs`, `.volumeMounts`) for all available containers and then filter that collection by the `ServiceBinding` `.spec.application.containers` filter before applying the appropriate modification.

If a `ClusterApplicationResourceMapping` defines `.envs` and `.volumeMounts`, the reconciler **MUST** first resolve a set of candidate locations in the application resource addressed by the `ServiceBinding` for all available containers and then filter that collection by the `ServiceBinding` `.spec.application.containers` filter before applying the appropriate modification.

If a `ServiceBinding` specifies a `.spec.applications.containers` value, and the value contains an `Int`-based index, that index **MUST** be used to filter the first entry in the `.containers` list and all other entries in those lists are ineligible for mapping.  If a `ServiceBinding` specifies a `.spec.applications.containers` value, and the value contains an `string`-based index that index **MUST** be used to filter all entries in the `.containers` list.  If a `ServiceBinding` specifies a `.spec.applications.containers` value and `ClusterApplicationResourceMapping` for the mapped type defines `.envs` and `.volumeMounts`, the reconciler **MUST** fail to reconcile.

A reconciler **MUST** apply the appropriate modification to the application resource addressed by the `ServiceBinding` as defined by `.volumes`.

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

    ```yaml
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
      - service.binding:timeout:sourceKey=db_timeout
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

Cluster operators and CRD authors **MAY** opt-in resources to service binding by defining a `ClusterRole` with a label matching `service.binding/controller=true`. For Provisioned Service resources the `get`, `list`, and `watch` verbs **MUST** be granted. For Application resources resources the `get`, `list`, `watch`, `update`, and `patch` verbs **MUST** be granted.

#### Example Resource

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: awesome-service-bindings
  labels:
    service.binding/controller: "true" # matches the aggregation rule selector
rules:
# for Provisioned Service resources only
- apiGroups:
  - awesome.example.com
  resources:
  - awesomeservices
  verbs:
  - get
  - list
  - watch
# for Application resources (also compatible with Provisioned Service resources)
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
