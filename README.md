# Service Binding RFC

RFC for binding services to runtime applications running in Kubernetes.  

## Terminology definition

Define common terms used in this domain.

## Motivation

*  Need a consistent way to bind k8s application to services (applications, databases, event streams, etc)
*  A standard / spec / RFC will enable adoption from different service providers
*  Cloud Foundry has done this well.  The equivalent is not available for k8s.

## Proposal

Main section of the doc.  Has sub-section that outline the design.

### Making a service bindable

*  What metadata is needed for a service to be bindable (e.g. name of fields it supports, name of secret it produces, etc)
*  Where is this information located

### Service Binding Schema

*  What is the schema of a Service Binding object / structure.
*  What metadata goes along that describes it

### Requesting binding

*  How do we request a binding from a service (assume the service has been provisioned)

### Injecting binding

*  Where in the container do we mount the binding information (e.g. what is the structure of the folders / files)
*  Consideration with clusters, namespaces, or VMs

### Extra:  Consuming binding

*  How are application expected to consume binding information 
*  Each framework may take a different approach, so this is about samples & recommendations (best practices)
*  Validates the design
