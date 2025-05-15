# Console for Kafka (cfk) - Technical Design Document

## 1. Executive Summary
Console for Kafka (cfk) is a terminal-based management console designed to simplify Apache Kafka operations. This document outlines the technical design for cfk, including the chosen technology stack, system architecture, API design, database schema, security considerations, and deployment strategies. The primary goal is to create a performant, user-friendly, and secure TUI application for Kafka management, built primarily using Go and the Bubble Tea TUI library.

## 2. Technology Stack

- **Programming Language:** Go
  - **Justification:** Excellent performance, strong concurrency support, cross-compilation to single binaries, rich ecosystem for CLI and network applications. Ideal for a responsive TUI application.
- **Terminal UI (TUI) Framework:** Bubble Tea (Go)
  - **Justification:** A powerful, Elm-inspired framework for building sophisticated terminal applications in Go. Offers flexibility and a good model for managing state and UI components.
- **Kafka Client Library:** `segmentio/kafka-go` (Pure Go) or `confluentinc/confluent-kafka-go` (CGO-based, official Confluent library)
  - **Justification:** `segmentio/kafka-go` offers a pure Go implementation, simplifying builds. `confluent-kafka-go` wraps `librdkafka`, providing extensive features and high performance but adds CGO dependency. Choice depends on balancing ease of build vs. feature set/performance needs. Initial recommendation: `segmentio/kafka-go` for simplicity, evaluate `confluent-kafka-go` if specific advanced features or performance bottlenecks arise.
- **Database:** None (for configuration management)
  - **Justification:** All configuration and cluster profiles are managed via configuration files (JSON, YAML, TOML) using Viper. No database is used for configuration or settings. Only audit logs may use a persistent file-based store if required.
- **Configuration Management:** `spf13/viper` (Go)
  - **Justification:** Handles various configuration file formats (JSON, YAML, TOML), environment variables, and command-line flags. Manages cfk's own settings and user-defined cluster connection profiles using configuration files, eliminating the need for a database for configuration.
- **CLI Argument Parsing (Optional):** `spf13/cobra` (Go)
  - **Justification:** While cfk is primarily TUI-driven, Cobra can be used for initial setup commands, configuration management outside the TUI, or non-interactive modes if needed in the future.
- **Logging:** `rs/zerolog` (Go)
  - **Justification:** High-performance, structured JSON logger. Useful for debugging and operational logging.
- **Encryption (for credentials):** OS-specific keychain libraries (e.g., `zalando/go-keyring` or `99designs/keyring`) or `golang.org/x/crypto/nacl/secretbox`.
  - **Justification:** Securely store sensitive data like Kafka SASL passwords. OS keychain is preferred for better integration; `secretbox` is a fallback for a self-managed encrypted store.

## 3. System Architecture

```mermaid
graph TD
    User[User Interaction] --> TUI_Layer[TUI Layer (Bubble Tea)]
    TUI_Layer --> Core_Logic[Core Logic Layer]
    Core_Logic --> Kafka_Client_Adapter[Kafka Client Adapter]
    Core_Logic --> Config_Manager[Configuration Manager]
    Core_Logic --> Audit_Logger[Audit Logger]
    Kafka_Client_Adapter --> Kafka_Cluster[Apache Kafka Cluster]
    Config_Manager --> Config_Files[(Config Files)]
    Audit_Logger --> Audit_Log_Store[(Audit Log Store)]
```

- **TUI Layer (Bubble Tea):**
  - Responsible for rendering all UI components (views, lists, forms, popups).
  - Captures user input (keyboard events).
  - Sends commands/messages to the Core Logic Layer based on user actions.
  - Receives updates from the Core Logic Layer to refresh the UI.
- **Core Logic Layer:**
  - Contains the main application business logic.
  - Manages application state (current view, selected items, cluster connections).
  - Orchestrates operations by interacting with other components (Kafka Client Adapter, Config Manager, Audit Logger).
  - Processes user commands from the TUI and translates them into Kafka operations or configuration changes.
- **Kafka Client Adapter:**
  - Abstracts the chosen Kafka client library (`segmentio/kafka-go` or `confluent-kafka-go`).
  - Provides a simplified interface for Kafka operations (connect, list topics, produce/consume messages, etc.) to the Core Logic Layer.
  - Handles connection management, error handling, and data marshalling/unmarshalling.
- **Configuration Manager:**
  - Manages cfk's application settings and user-defined Kafka cluster configurations.
  - Reads from and writes to configuration files via Viper (no database used).
  - Handles encryption/decryption of sensitive configuration values (e.g., passwords).
- **Audit Logger:**
  - Logs significant user actions (e.g., creating/deleting topics, resetting offsets) to the SQLite database for auditing purposes.
- **Data Flow:**
  1. User interacts with the TUI (e.g., presses a key to list topics).
  2. TUI Layer translates this into a command for the Core Logic Layer.
  3. Core Logic Layer processes the command, potentially retrieving cluster configuration via Config Manager.
  4. Core Logic Layer instructs Kafka Client Adapter to perform the Kafka operation (e.g., fetch topic list).
  5. Kafka Client Adapter communicates with the Kafka Cluster.
  6. Kafka Cluster responds to Kafka Client Adapter.
  7. Kafka Client Adapter returns data/status to Core Logic Layer.
  8. Core Logic Layer updates its state and sends an update message to the TUI Layer.
  9. TUI Layer re-renders the relevant parts of the UI to display the new information.
  10. If the action is auditable, Core Logic Layer instructs Audit Logger to record the event.

## 4. API Design (Internal Interfaces)

This section describes the primary internal Go interfaces or function signatures for interacting with Kafka and managing cfk components. These are not HTTP APIs.

**Cluster Management:**
- `ConnectCluster(config ClusterDetails) (ClusterSession, error)`
- `DisconnectCluster(session ClusterSession) error`
- `GetClusterMetadata(session ClusterSession) (ClusterMetadata, error)`
- `ListBrokers(session ClusterSession) ([]BrokerInfo, error)`

**Topic Management:**
- `ListTopics(session ClusterSession) ([]TopicSummary, error)`
- `GetTopicDetails(session ClusterSession, topicName string) (TopicDetails, error)`
- `CreateTopic(session ClusterSession, config CreateTopicConfig) error`
- `DeleteTopic(session ClusterSession, topicName string) error`
- `UpdateTopicConfig(session ClusterSession, topicName string, configs map[string]string) error`

**Consumer Group Management:**
- `ListConsumerGroups(session ClusterSession) ([]ConsumerGroupSummary, error)`
- `GetConsumerGroupDetails(session ClusterSession, groupID string) (ConsumerGroupDetails, error)`
- `ResetConsumerGroupOffsets(session ClusterSession, groupID string, resetParams OffsetResetParameters) error`
- `DeleteConsumerGroup(session ClusterSession, groupID string) error`

**Message Handling:**
- `ProduceMessage(session ClusterSession, topicName string, partition int32, key []byte, value []byte, headers map[string][]byte) error`
- `ConsumeMessages(session ClusterSession, topicName string, partition int32, offset int64, maxMessages int, filter MessageFilter) (<-chan KafkaMessage, error)`

**Configuration API (ConfigManager):**
- `LoadClusterConfigs() ([]ClusterProfile, error)`
- `SaveClusterConfig(profile ClusterProfile) error`
- `DeleteClusterConfig(profileID string) error`
- `GetAppSetting(key string) (string, error)`
- `SetAppSetting(key string, value string) error`

**Audit Log API (AuditLogger):**
- `LogAction(actor string, action string, resourceType string, resourceName string, details map[string]interface{}) error`

## 5. Configuration and Audit Log Storage Design

**Configuration Storage:**
- All cluster profiles and application settings are stored in configuration files (JSON, YAML, or TOML) managed by Viper. No database is used for configuration or settings.
- Sensitive values (e.g., passwords) are encrypted before being written to config files.

**Audit Log Storage:**
- Audit logs are stored in a persistent file-based store (e.g., a local SQLite file or structured log file). Only audit logs require persistent storage; configuration does not.
## 6. Security Considerations

- **Secure Credential Storage:**
  - Kafka passwords and sensitive connection parameters stored in `cluster_profiles.sasl_password_encrypted` will be encrypted using OS keychain services (via `zalando/go-keyring` or similar) or, if unavailable, a strong symmetric encryption algorithm (e.g., AES-GCM with a key derived from a user-provided master password or a securely stored local key).
- **TLS/SSL Connections:**
  - Full support for TLS/SSL encrypted connections to Kafka brokers, including configuration for CA certificates, client certificates, and key files. Option to skip TLS verification (with warnings) for development/testing environments.
- **Kafka Authentication:**
  - Support for SASL mechanisms (PLAIN, SCRAM-SHA-256/512, GSSAPI/Kerberos where `librdkafka` is used and configured).
  - Future: OAuth/OIDC integration for Kafka clusters supporting it (may require external browser interaction for token acquisition).
- **Role-Based Access Control (RBAC):**
  - cfk will respect Kafka's native ACLs. Operations will succeed or fail based on the connected user's permissions on the Kafka cluster.
  - For cfk's own features (if any local RBAC is implemented for shared cfk instances, not typical for a personal CLI tool), it would be managed via separate configuration.
- **Audit Logging:**
  - Comprehensive audit logs (as per `audit_logs` table design) will record significant actions performed through cfk, providing traceability.
- **Input Validation:**
  - All user inputs used in constructing Kafka requests or configurations will be validated to prevent unintended behavior. However, the primary risk here is misconfiguration rather than typical injection vulnerabilities found in web apps.
- **Dependency Security:**
  - Regularly scan Go module dependencies for known vulnerabilities using tools like `govulncheck`.

## 7. Scalability and Performance

- **Asynchronous Operations:** All Kafka interactions (fetching data, producing messages) will be performed asynchronously to prevent blocking the TUI and maintain responsiveness.
- **Efficient Data Handling:**
  - Paginate large lists (topics, consumer groups, messages).
  - Fetch only necessary data for the current view.
  - Use efficient data structures and algorithms in the Go backend.
- **Optimized TUI Rendering:**
  - Leverage Bubble Tea's capabilities for efficient rendering and updates.
  - Only re-render components that have changed.
- **Connection Pooling (if applicable):** While typically a single connection per managed cluster is used, ensure efficient management of these connections.
- **Resource Management:** Proper closing of Kafka connections, file handles, and goroutines to prevent leaks.
- **Lazy Loading:** Load detailed information (e.g., topic configurations, full message content) on demand rather than all at once.

## 8. Development and Deployment

- **Development Practices:**
  - **Version Control:** Git, hosted on a platform like GitHub/GitLab.
  - **Branching Strategy:** Gitflow or GitHub Flow.
  - **Code Formatting & Linting:** `gofmt`, `golangci-lint`.
  - **Testing:**
    - Unit tests for core logic, utility functions.
    - Integration tests for Kafka client adapter interactions (using a local Kafka instance like `testcontainers-go` or a mocked Kafka).
    - TUI interaction testing might be manual initially, explore TUI testing tools if available/practical.
  - **Continuous Integration (CI):** GitHub Actions or similar to automate builds, tests, linting on every push/PR.
- **Deployment Strategy:**
  - **Distribution:** Single, statically-linked executable binaries for Linux, macOS, and Windows.
  - **Release Management:** Use Git tags for versioning. Automated releases via CI (e.g., using GoReleaser).
  - **Packaging (Optional):** Consider Homebrew, Scoop, Snap, or .deb/.rpm packages for easier installation.
  - **Docker Images:** Provide official Docker images on Docker Hub for containerized usage.
- **Environment Setup:**
  - Document required Go version and any build-time dependencies.
  - Instructions for setting up a local Kafka environment for development and testing.

## 9. Open Source Projects and Third-party Services

- **Go:** Core language (golang.org)
- **Bubble Tea:** TUI framework (github.com/charmbracelet/bubbletea)
- **Lip Gloss:** Styling for Bubble Tea (github.com/charmbracelet/lipgloss)
- **Kafka Client:** `segmentio/kafka-go` or `confluentinc/confluent-kafka-go`

- **Configuration:** `spf13/viper`
- **CLI (Optional):** `spf13/cobra`
- **Logging:** `rs/zerolog`
- **Encryption/Keychain:** `zalando/go-keyring` or `99designs/keyring`, `golang.org/x/crypto`
- **Testing:** `testify/assert`, `testify/mock`, `testcontainers-go`
- **CI/CD:** GitHub Actions (or similar)
- **Release Automation:** GoReleaser (goreleaser.com)

**Justification:** These libraries are well-maintained, widely used in the Go community, and provide the necessary functionalities for building cfk as per the requirements.

## 10. Risks and Mitigations

- **Risk: TUI Complexity & Responsiveness**
  - **Description:** Building a feature-rich and responsive TUI like k9s is challenging. Performance issues can arise with large amounts of data or complex views.
  - **Mitigation:** Use Bubble Tea effectively, implement asynchronous operations for all I/O, optimize rendering, employ pagination and virtual scrolling, conduct performance testing early.
- **Risk: Kafka Client Integration & Compatibility**
  - **Description:** Supporting various Kafka versions, security configurations (SASL, SSL), and handling diverse error conditions can be complex.
  - **Mitigation:** Choose a mature Kafka client library. Thoroughly test against different Kafka versions and security setups. Provide clear documentation for connection configurations.
- **Risk: Secure Credential Management**
  - **Description:** Improper handling of Kafka credentials can lead to security vulnerabilities.
  - **Mitigation:** Implement robust encryption for stored credentials using OS keychain or strong cryptographic libraries. Avoid logging sensitive information. Educate users on best practices.
- **Risk: Cross-Platform TUI Behavior**
  - **Description:** Terminal emulators behave differently across OS (Windows, macOS, Linux), potentially leading to rendering issues or inconsistent behavior.
  - **Mitigation:** Test extensively on all target platforms. Rely on established TUI libraries that handle cross-platform differences. Provide clear guidance on recommended terminal emulators.
- **Risk: Scope Creep**
  - **Description:** The feature set can expand beyond the initial MVP, delaying releases.
  - **Mitigation:** Adhere to the defined roadmap and prioritize features based on user value and development effort. Clearly define MVP scope.

## 11. Next Steps

1.  **Setup Project:** Initialize Go module, Git repository, basic project structure.
2.  **Proof of Concept (PoC) - Basic TUI & Kafka Connection:**
    *   Implement a minimal Bubble Tea application.
    *   Integrate chosen Kafka client to connect to a Kafka cluster and list topics.
    *   Basic configuration loading for cluster connection details.
3.  **MVP Development (Milestone 1 from Product Doc):**
    *   **Cluster Management:** Add/edit/remove cluster configurations (stored in SQLite), view cluster health/broker status.
    *   **Topic Management:** List, create, delete topics. View topic details (partitions, replication).
    *   **Consumer Group Management:** List consumer groups, view lag/offsets.
    *   **Message Inspection:** Basic message browsing (consuming latest messages from a topic).
    *   **Message Production:** Simple interface to produce messages to a topic.
    *   **Basic Authentication:** Support for SASL PLAIN/SCRAM in cluster configurations.
    *   **Database Setup:** Implement SQLite schema for `cluster_profiles` and `app_settings`.
4.  **Milestone 2 (Monitoring & Advanced Search):**
    *   Develop monitoring dashboard components (real-time metrics).
    *   Implement in-terminal alerts/notifications.
    *   Add advanced message search and filtering capabilities.
5.  **Milestone 3 (Security & Multi-Cluster):**
    *   Implement RBAC considerations (respecting Kafka ACLs).
    *   Implement `audit_logs` table and logging functionality.
    *   Enhance multi-cluster support and navigation.
    *   Add troubleshooting tools (e.g., partition reassignment views).
6.  **Documentation & Testing:** Continuously write user and developer documentation. Expand test coverage.
7.  **Release Early, Release Often:** Consider alpha/beta releases to gather user feedback.