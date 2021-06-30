# User Guide

## Introduction

**TODO**

## Installing Service Binding Controller

**TODO**

## Creating Provisioned Services

**TODO**

## Binding Workloads with Services

**TODO**

## Consuming the Bindings from Workloads

The [Workload Projection section](https://github.com/k8s-service-bindings/spec#workload-projection) of the specification describes how bindings are projected into the workload.  The primary mechanism of projection is through files mounted at a specific directory.  The bindings directory path is discovered through the mandatory `$SERVICE_BINDING_ROOT` environment variable set on all containers where bindings are created.

Within this service binding root directory, multiple Service Bindings may be projected.  For example, a workload that requires both a database and event stream will declare one `ServiceBinding` for the database, a second `ServiceBinding` for the event stream, and both bindings will be projected as subdirectories of the root.

Let's take a look at the example given in the spec:

```
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
In the above example, there are two bindings under the `$SERVICE_BINDING_ROOT` directory with the names `account-database` and `transaction-event-stream`.  In order for a workload to configure itself, it must select the proper binding for each client type.  Each binding directory has a special file named `type` and you can use the content of that file to identify the type of the binding projected into that directory (e.g. `mysql`, `kafka`).  Some bindings optionally also contain another special file named `provider` which is an additional identifier used to further narrow down ambiguous types.  Choosing a binding "by-name" is not considered good practice as it makes your workload less portable (although it may be unavoidable).  Wherever possible use the `type` field and, if necessary, `provider` to select a binding.

Usually, operators use `ServiceBinding` resource name (`.metadata.name`) as the bindings directory name, but the spec also provides a way to override that name through the `.spec.name` field. So, there is a chance for bindings name collision.  However, due to the nature of the volume mount in Kubernetes, the bindings directory will contain values from only one of the Secret resources.  This is a caveat of using the bindings directory name to look up the bindings.

### Purpose of the type and the provider fields in the workload projection

The specification mandates the `type` field and recommends `provider` field in the projected binding.  In many cases the `type` field should be good enough to select the appropriate binding.  In cases where it is not (e.g. when there are different providers for the same Provisioned Service), the `provider` field may be used.  For example, when the type is `mysql`, the `provider` value might be `mariadb`, `oracle`, `bitnami`, `aws-rds`, etc.  When the workload is selecting a binding, if necessary, it could consider `type` and `provider` as a composite key to avoid ambiguity.  This could be helpful if a workload needs to choose a particular provider based on the deployment environment.  In the deployment environment (`stage`, `prod`, etc.), at any given time, you need to ensure only one binding projection exists for a given `type` or `type` and `provider` -- unless your workload needs to connect to all the services.

### Programming language specific library APIs

A workload can retrieve bindings through a library available for your programming language of choice.  Language-specific APIs are encouraged to follow the pattern described here, but may not.  Consult your library API documentation to confirm its behavior.

For languages such as Go without operator overloading, separate functions can be used to retrieve bindings:

```
Bindings(_type string) []Binding
BindingsWithProvider(_type, provider string) []Binding
```

For languages such as Java with operator overloading, the same method name with different argument lists can be used to retrieve bindings:

```
public List<Binding> filterBindings(@Nullable String type)
public List<Binding> filterBindings(@Nullable String type, @Nullable String provider)
```

(Example taken from [Spring Cloud Bindings](https://github.com/spring-cloud/spring-cloud-bindings))

The specification does not guarantee a single binding of a given type or type & provider tuple so the APIs return collections of bindings.  Depending on your workload need, you can choose to connect to the first entry or all of them.

### Environment Variables

The specification also has support for projecting binding values as environment variables.  You can use the built-in language feature of your programming language of choice to read environment variables.  The container must restart to update the values of environment variables if there is a change in the binding.
