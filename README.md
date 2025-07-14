# MessageFlow

[![Run Tests](https://github.com/denchenko/messageflow/actions/workflows/go.yml/badge.svg?branch=main)](https://github.com/denchenko/messageflow/actions/workflows/go.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/denchenko/messageflow)](https://goreportcard.com/report/github.com/denchenko/messageflow)
[![GoDoc](https://godoc.org/github.com/denchenko/messageflow?status.svg)](https://godoc.org/github.com/denchenko/messageflow)

MessageFlow is a Go library and CLI tool for visualizing AsyncAPI specifications. It provides tools to parse AsyncAPI documents and transform them into visual formats, making it easier to understand message flows and service interactions in asynchronous systems.

Example of visualizing a Notification service using [this](pkg/schema/source/asyncapi/testdata/notification.yaml) AsyncAPI specification. It can be useful to display service communication with a message bus without requiring detailed knowledge about other services in the ecosystem. Message payloads are displayed as thumbnails when hovering over specific queues. This approach was chosen to keep the schema clean and uncluttered.

![schema](pkg/schema/target/d2/testdata/service_channels_notification.svg)

When you have AsyncAPI specifications for all services in your system, MessageFlow can generate comprehensive documentation showing the complete service ecosystem. See [examples/docs](examples/docs) for a complete multi-service documentation example. For instance, in the generated documentation, the same service now appears like this:

![schema](examples/docs/diagrams/service_notification-service.svg)

## Usage

MessageFlow provides a command-line interface.

```bash
go install github.com/denchenko/messageflow/cmd/messageflow@latest
```

### Generate Schema

The `gen-schema` command processes AsyncAPI files and generates formatted schemas or rendered diagrams:

```bash
# Generate and render a diagram
messageflow gen-schema --target d2 --render-to-file schema.svg --asyncapi-files asyncapi.yaml

# Generate formatted schema only
messageflow gen-schema --format-to-file schema.d2 --asyncapi-files asyncapi.yaml

# Process multiple AsyncAPI files
messageflow gen-schema --render-to-file combined.svg --asyncapi-files "file1.yaml,file2.yaml,file3.yaml"
```

### Generate Documentation

The `gen-docs` command generates comprehensive markdown documentation from AsyncAPI files, including diagrams and changelog tracking:

```bash
# Generate documentation for multiple services
messageflow gen-docs --asyncapi-files "service1.yaml,service2.yaml,service3.yaml" --output ./docs
```

The generated documentation includes:
- **Context diagram**: Overview of all services and their interactions
- **Service diagrams**: Individual diagrams showing each service's channels and operations
- **Channel diagrams**: Detailed views of message flows through specific channels
- **Changelog tracking**: Automatic detection and documentation of schema changes between runs
- **Message payloads**: JSON schemas for all message types


## Known Limitations

* One kind of message per channel;
* One kind of server protocol per spec.
