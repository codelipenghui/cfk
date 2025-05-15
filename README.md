# Console for Kafka (cfk)

A terminal-based management console designed to simplify Apache Kafka operations.

## Overview

Console for Kafka (cfk) is a TUI (Terminal User Interface) application that allows you to manage and interact with Apache Kafka clusters directly from your terminal. It provides a user-friendly interface for common Kafka operations such as browsing topics, viewing messages, and managing consumer groups.

## Features

- Connect to multiple Kafka clusters
- Browse and search topics
- View topic details (partitions, replication factor)
- Produce and consume messages
- Monitor consumer groups
- Support for authentication (SASL PLAIN, SCRAM)

## Installation

### Prerequisites

- Go 1.21 or later

### Building from source

```bash
# Clone the repository
git clone https://github.com/cfk-dev/cfk.git
cd cfk

# Build the application
go build -o cfk ./cmd/cfk

# Install the application (optional)
go install ./cmd/cfk
```

## Usage

Simply run the `cfk` command to start the application:

```bash
./cfk
```

On first run, a default configuration file will be created at `~/.cfk/config.yaml`. You can edit this file to add your Kafka cluster configurations.

### Configuration

The configuration file is located at `~/.cfk/config.yaml` and has the following structure:

```yaml
clusters:
  - name: local-kafka
    bootstrap_servers:
      - localhost:9092
    ssl: false
    sasl: false
  - name: secured-kafka
    bootstrap_servers:
      - kafka.example.com:9093
    username: user
    password: password
    ssl: true
    sasl: true
    sasl_type: PLAIN
ui:
  theme: default
  refresh_interval: 5
  max_messages_shown: 100
```

## Project Structure

```
.
├── cmd/
│   └── cfk/            # Main application entry point
├── internal/
│   ├── config/         # Configuration management
│   ├── core/           # Application core logic
│   ├── kafka/          # Kafka client adapter
│   └── tui/            # Terminal UI components
├── go.mod              # Go module definition
└── README.md           # This file
```

## License

MIT