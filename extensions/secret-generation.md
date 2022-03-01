# Secret Generation Extension

This document defines an extension to the [Service Binding Specification for Kubernetes](https://github.com/servicebinding/spec) ("Service Binding spec" for short henceforth).  This extension specifies generating a Kubernetes Secret resource that can be consumed by any Service Binding spec compliant implementation.  The Secret resource is generated from one of these sources:

- Operator Lifecycle Manager Descriptors
- Custom Resource Definition Annotations
- Custom Resource Annotations

## Status

This document is a pre-release, working draft of the Secret Generation extension for Service Binding, representing the collective efforts of the community.  It is published for early implementors and users to provide feedback.  Any part of this document may change before the extension reaches 1.0 with no promise of backwards compatibility.

## Specification

Many services, especially initially, will not be Provisioned Service-compliant.  These services will expose the appropriate binding `Secret` information, but not in the way that the specification or applications expect.  Users should have a way of describing a mapping from existing data associated with arbitrary resources and CRDs to a representation of a binding `Secret`.

To handle the majority of existing resources and CRDs, `Secret` generation needs to support the following behaviors:

1.  Extract a string from a resource
1.  Extract an entire `ConfigMap`/`Secret` refrenced from a resource
1.  Extract a specific entry in a `ConfigMap`/`Secret` referenced from a resource
1.  Extract entries from a collection of objects, mapping keys and values from entries in a `ConfigMap`/`Secret` referenced from a resource
1.  Exctact a collection of specific entry values in a resource's collection of objects
1.  Map each value to a specific key
1.  Map each value of a collection to a key with generated name

While the syntax of the generation strategies are specific to the system they are annotating, they are based on a common data model.

| Model | Description
| ----- | -----------
| `path` | A template represention of the path to an element in a Kubernetes resource.  The value of `path` is specified as [JSONPath](https://kubernetes.io/docs/reference/kubectl/jsonpath/).  Required.
| `objectType` | Specifies the type of the object selected by the `path`.  One of `ConfigMap`, `Secret`, or `string` (default).
| `elementType` | Specifies the type of object in an array selected by the `path`.  One of `sliceOfMaps`, `sliceOfStrings`, `string` (default).
| `sourceKey` | Specifies a particular key to select if a `ConfigMap` or `Secret` is selected by the `path`.  Specifies a value to use for the key for an entry in a binding `Secret` when `elementType` is `sliceOfMaps`.
| `sourceValue` | Specifies a particular value to use for the value for an entry in a binding `Secret` when `elementType` is `sliceOfMaps` or `sliceOfStrings`.


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
  tags:
  - Brno
  - PWR
  - stage
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

1. Mount the items of a collection into the binding `Secret` with one key per item

    ```yaml
    - path: spec.tags
      x-descriptors:
      - service.binding:tags:elementType=sliceOfStrings
    ```

1. Mount the values of collection entries into the binding `Secret` with one key per entry value

    ```yaml
   - path: bootstrap
      x-descriptors:
      - service.binding:endpoints:elementType=sliceOfStrings:sourceValue=url
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
  tags:
  - Brno
  - PWR
  - stage
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
1. Mount the items of a collection into the binding `Secret` with one key per item
    ```plain
    "service.binding/tags":
      "path={.spec.tags},elementType=sliceOfStrings
    ```
1. Mount the values of collection entries into the binding `Secret` with one key per entry value
    ```plain
    “service.binding/endpoints”:
      "path={.status.bootstrap},elementType=sliceOfStrings,sourceValue=url"
    ```
