![scim-logo](LOGO.png)

[![GoVersion](https://img.shields.io/github/go-mod/go-version/elimity-com/scim.svg)](https://github.com/elimity-com/scim)
[![GoDoc](https://img.shields.io/badge/godoc-reference-blue.svg)](https://pkg.go.dev/github.com/elimity-com/scim)


[![Tag](https://img.shields.io/github/tag/elimity-com/scim.svg)](https://gitHub.com/elimity-com/scim/releases)

This is an open source implementation of the [SCIM v2.0](http://www.simplecloud.info/#Specification) specification for use in Golang.
SCIM defines a flexible schema mechanism and REST API for managing identity data.
The goal is to reduce the complexity of user management operations by providing patterns for exchanging schemas using HTTP.

In this implementation it is easy to add *custom* schemas and extensions with the provided structures.
Incoming resources will be *validated* by their corresponding schemas before being passed on to their callbacks.

The following features are supported:
- GET for `/Schemas`, `/ServiceProviderConfig` and `/ResourceTypes`
- CRUD (POST/GET/PUT/DELETE and PATCH) for your own resource types (i.e. `/Users`, `/Groups`, `/Employees`, ...)

Other optional features such as sorting, bulk, etc. are **not** supported in this version.

## Installation
Assuming you already have a (recent) version of Go installed, you can get the code with go get:
```bash
$ go get github.com/elimity-com/scim
```

## Usage
**!** errors are ignored for simplicity.
### 1. Create a service provider configuration.
[RFC Config](https://tools.ietf.org/html/rfc7643#section-5) |
[Example Config](https://tools.ietf.org/html/rfc7643#section-8.5)
```go
config := scim.ServiceProviderConfig{
    DocumentationURI: optional.NewString("www.example.com/scim"),
}
```
**!** no additional features/operations are supported in this version.

### 2. Create all supported schemas and extensions.
[RFC Schema](https://tools.ietf.org/html/rfc7643#section-2) |
[User Schema](https://tools.ietf.org/html/rfc7643#section-4.1) |
[Group Schema](https://tools.ietf.org/html/rfc7643#section-4.2) |
[Extension Schema](https://tools.ietf.org/html/rfc7643#section-4.3)
```go
schema := schema.Schema{
    ID:          "urn:ietf:params:scim:schemas:core:2.0:User",
    Name:        optional.NewString("User"),
    Description: optional.NewString("User Account"),
    Attributes:  []schema.CoreAttribute{
        schema.SimpleCoreAttribute(schema.SimpleStringParams(schema.StringParams{
            Name:       "userName",
            Required:   true,
            Uniqueness: schema.AttributeUniquenessServer(),
        })),
    },
}

extension := schema.Schema{
    ID:          "urn:ietf:params:scim:schemas:extension:enterprise:2.0:User",
    Name:        optional.NewString("EnterpriseUser"),
    Description: optional.NewString("Enterprise User"),
    Attributes: []schema.CoreAttribute{
        schema.SimpleCoreAttribute(schema.SimpleStringParams(schema.StringParams{
            Name: "employeeNumber",
        })),
        schema.SimpleCoreAttribute(schema.SimpleStringParams(schema.StringParams{
            Name: "organization",
        })),
    },
}
```

### 3. Create all resource types and their callbacks.
[RFC Resource Type](https://tools.ietf.org/html/rfc7643#section-6) |
[Example Resource Type](https://tools.ietf.org/html/rfc7643#section-8.6)

#### 3.1 Callback (implementation of `ResourceHandler`)
[Simple In Memory Example](resource_handler_test.go)
```go
var userResourceHandler scim.ResourceHandler
// initialize w/ own implementation
```
**!** each resource type should have its own resource handler.

#### 3.2 Resource Type
```go
resourceTypes := []ResourceType{
    {
        ID:          optional.NewString("User"),
        Name:        "User",
        Endpoint:    "/Users",
        Description: optional.NewString("User Account"),
        Schema:      schema,
        SchemaExtensions: []SchemaExtension{
            {Schema: extension},
        },
        Handler:     userResourceHandler,
    },
},
```

### 4. Create Server
```go
server := Server{
    Config:        config,
    ResourceTypes: resourceTypes,
}
```

## Addition Checks/Tests
Not everything can be checked by the SCIM server itself.
Below are some things listed that we expect that the implementation covers.

**!** this list is currently incomplete!

We want to keep this list as short as possible. 
If you have ideas how we could enforce these rules in the server itself do not hesitate to open
[an issue](https://github.com/elimity-com/scim/issues/new) or a PR.
### Mutability
#### Immutable Attributes
*PUT Handler*: If one or more values are already set for the attribute, the input value(s) MUST match.
#### WriteOnly Attributes
*ALL Handlers*: Attribute values SHALL NOT be returned. \
Note: These attributes usually also has a returned setting of "never".

## Contributing
[![Contributors](https://img.shields.io/github/contributors/elimity-com/scim.svg)](https://gitHub.com/elimity-com/scim/contributors/)

We are happy to review pull requests,
but please first discuss the change you wish to make via issue, email,
or any other method with the owners of this repository before making a change.

If you would like to propose a change please ensure the following:
- All checks of GitHub Actions are passing ([GolangCI-Lint](https://github.com/golangci/golangci-lint): `misspell`, `godot` and `whitespace`)
- All already existing tests are passing.
- You have written tests that cover the code you are making, make sure to include edge cases.
- There is documentation for at least all public functions you have added.
- New public functions and structures are kept to a minimum.
- The same practices are applied (such as the anatomy of methods, names, etc.)
- Your changes are compliant with SCIM v2.0 (released as
[RFC7642](https://tools.ietf.org/html/rfc7642),
[RFC7643](https://tools.ietf.org/html/rfc7643) and
[RFC7644](https://tools.ietf.org/html/rfc7644) under [IETF](https://ietf.org/)).
