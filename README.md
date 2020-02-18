# Service Binding Specification

Specification for binding services to runtime applications running in Kubernetes.  

## Terminology definition

*  **service** - any software that is exposing functionality.  Could be a RESTful application, a database, an event stream, etc.
*  **application** - in this specification we refer to a single runtime-based microservice (e.g. MicroProfile app, or Node Express app) as an application.  This is different than an umbrella (SIG) _Application_ which refers to a set of microservices.
*  **binding** - providing the necessary information for an application to connect to a service.
*  **secret** - refers to a Kubernetes [Secret](https://kubernetes.io/docs/concepts/configuration/secret/).
*  **config map** - refers to a Kubernetes [ConfigMap](https://kubernetes.io/docs/tasks/configure-pod-container/configure-pod-configmap/).

## Motivation

*  Need a consistent way to bind k8s application to services (applications, databases, event streams, etc)
*  A standard / spec / RFC will enable adoption from different service providers
*  Cloud Foundry has done this well.  The equivalent is not available for k8s.

## Proposal

Main section of the doc.  Has sub-section that outline the design.

### 1.  Making a service bindable

This specification aims to enable the following four scenarios of bindable services:
1. Services deployed via an OLM-backed Operator
1. Services deployed via an Operator (without OLM)
1. Services deployed directly to k8s / OpenShift (using a k8s Deployment)
1. Services deployed outside of k8s (in a VM or via a Cloud service)

For a service to be bindable it **should** provide:
* a ConfigMap that contains the name (or pattern) of the Secret holding the binding data, and describes metadata associated with each of the items referenced in the Secret.  

For a service to be bindable it **must** provide:
* a Secret that contains the binding data (see section #2)
* a reference in one of its deployable resources that points to either its ConfigMap (recommended) or Secret (minimum).

The reference's location and format depends on the scenarios (1-4 above)

1. OLM-based - Using a statusDescriptor in the CSV to hold the reference to the ConfigMap / Secret
  * x-descriptors:
    - urn:alm:descriptor:io.kubernetes:ConfigMap
    - servicebinding:ConfigMap
  * x-descriptors:
    - urn:alm:descriptor:io.kubernetes:Secret
    - servicebinding:Secret

2. Operator-based (without OLM) - An annotation in the Operator's CRD
  * The annotation is in the form of:
    * `servicebinding/configMap: <bindable-configmap>`
    * `servicebinding/secret: <bindable-secret>`

3. Deploymented-based (no Operator) - An annotation in the Deployment's CR
  * The annotation is in the form of:
    * `servicebinding/configMap: <bindable-configmap>`
    * `servicebinding/secret: <bindable-secret>`
    
4. External service - An annotation in the local ConfigMap or Secret that bridges the external service
  * The annotation is in the form of:
    * `servicebinding/configMap: self`
    * `servicebinding/secret: self`   


### 2.  Service Binding Schema

*  What is the schema of a Service Binding object / structure.
*  What metadata goes along that describes it

### 3.  Request service binding

*  How do we request a binding from a service (assume the service has been provisioned)
*  How is that binding authorized?

### 4.  Mounting binding information

*  Where in the container do we mount the binding information (e.g. what is the structure of the folders / files)
*  Consideration with clusters, namespaces, or VMs

### Extra:  Consuming binding

*  How are application expected to consume binding information 
*  Each framework may take a different approach, so this is about samples & recommendations (best practices)
*  Validates the design
