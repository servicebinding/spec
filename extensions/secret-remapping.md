# Secret Remapping Extension

This document defines an extension to the [Service Binding Specification for Kubernetes](https://github.com/servicebinding/spec) ("Service Binding spec" for short henceforth).  Many services are not Provisioned Service-compliant.  These are services that provides Secret resources that may not have the keys expected by the application.  Users should have a way of describing a mapping from current Secret keys to new Secret keys.  This extension specifies a mapping extension to remap a Secret resource keys.  This extension call the remapped resource conforming to the Provisioned Service as SecretRemapping.

## Status

This document is a pre-release, working draft of the Composite Service extension for Service Binding, representing the collective efforts of the community.  It is published for early implementors and users to provide feedback.  Any part of this document may change before the extension reaches 1.0 with no promise of backwards compatibility.

## Notational Conventions

The key words "MUST", "MUST NOT", "REQUIRED", "SHALL", "SHALL NOT", "SHOULD", "SHOULD NOT", "RECOMMENDED", "NOT RECOMMENDED", "MAY", and "OPTIONAL" in this document are to be interpreted as described in [BCP 14](https://tools.ietf.org/html/bcp14) [RFC2119](https://tools.ietf.org/html/rfc2119) [RFC8174](https://tools.ietf.org/html/rfc8174) when, and only when, they appear in all capitals, as shown here.

An implementation is not compliant if it fails to satisfy one or more of the MUST, MUST NOT, REQUIRED, SHALL, or SHALL NOT requirements for the protocols it implements.  An implementation is compliant if it satisfies all the MUST, MUST NOT, REQUIRED, SHALL, and SHALL NOT requirements for the protocols it implements.

## Specification

A SecretRemapping describes a mapping of Secret resource keys from current keys to new keys.  The `.spec.name` **MUST** refer to a Secret resource.  The `.spec.mapping` **MUST** be an array of of two strings where one of the string called `current` points to the current key and the other string called `new` points to the new key.

### Resource Type Schema

```yaml
apiVersion: extensions.servicebinding.io/v1alpha1
kind: SecretRemapping
metadata:
  name: # string
  generation: # int64, defined by the Kubernetes control plane
  ...
spec:
  name: # string
  mapping: # []Mapping
  - current: # string
    new: # string
status:
  binding: # LocalObjectReference, optional
    name: # string
  conditions: # []metav1.Condition containing at least one entry for `Ready`
  observedGeneration: # int64
```

### Example Resource

```
apiVersion: extensions.servicebinding.io/v1alpha1
kind: SecretRemapping
metadata:
  name: database-credentials
spec:
  name: db-secret-generated
  mapping:
  - current: user
    new: username
  - current: passwd
    new: password
status:
  binding:
    name: database-credentials-ylt42
  conditions:
  - lastTransitionTime: "2021-07-24T06:10:01Z"
    message: Secret resource created
    reason: SecretCreated
    status: "True"
    type: Ready
  observedGeneration: 1
```

### Reconciler Implementation

A Reconciler implementation for the `SecretRemapping` type is responsible for creating a Secret from an existing Secret with new keys.  The generated Secret resource name **MUST** be set in the `.status.binding.name` attribute to make the `SecretRemapping` resource conform to a Provisioned Service.

#### Ready Condition Status

If the Secret remapping is completed successfully, the `Ready` condition status **MUST** be set to `True`.  If the Secret remapping cannot be completed, the `Ready` condition status **MUST** be set to `False`.  If the `Ready` condition status is neither actively `True` nor `False` it **SHOULD** be set to `Unknown`.
