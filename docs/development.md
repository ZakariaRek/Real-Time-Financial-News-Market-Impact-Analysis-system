# 👨‍💻 Development Guide

<div align="center">

![Go](https://img.shields.io/badge/Go-1.23-00ADD8?style=for-the-badge&logo=go&logoColor=white)
![Java](https://img.shields.io/badge/Java-21-ED8B00?style=for-the-badge&logo=openjdk&logoColor=white)
![Git](https://img.shields.io/badge/Git-F05032?style=for-the-badge&logo=git&logoColor=white)
![VSCode](https://img.shields.io/badge/VS_Code-007ACC?style=for-the-badge&logo=visual-studio-code&logoColor=white)

**Development Guide for Financial Intelligence Platform**

[Setup](#-development-setup) • [Standards](#-coding-standards) • [Testing](#-testing) • [Debugging](#-debugging)

</div>

---

## 📋 Table of Contents

- [Development Setup](#-development-setup)
- [Project Structure](#-project-structure)
- [Coding Standards](#-coding-standards)
- [Git Workflow](#-git-workflow)
- [Testing](#-testing)
- [Debugging](#-debugging)
- [Code Generation](#-code-generation)
- [Database Migrations](#-database-migrations)
- [Performance Optimization](#-performance-optimization)
- [Contributing](#-contributing)

---

## 🚀 Development Setup

### Prerequisites

```mermaid
graph LR
    subgraph "Required Tools"
        GO[Go 1.23+]
        JAVA[Java 21+]
        DOCKER[Docker]
        GIT[Git]
    end

    subgraph "Optional Tools"
        VSCODE[VS Code]
        IDEA[IntelliJ IDEA]
        POSTMAN[Postman]
        PGADMIN[pgAdmin]
    end

    subgraph "Build Tools"
        MAVEN[Maven 3.9+]
        PROTOC[Protocol Buffers]
    end

    GO --> VSCODE
    JAVA --> IDEA
    DOCKER --> POSTMAN
    GIT --> VSCODE

    style GO fill:#00ADD8,stroke:#00758F,color:#fff
    style JAVA fill:#ED8B00,stroke:#B86A00,color:#fff
    style DOCKER fill:#2496ED,stroke:#1565C0,color:#fff
```

### Installation Steps

#### 1. Install Go

```bash
# Download and install Go 1.23
wget https://go.dev/dl/go1.23.0.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.23.0.linux-amd64.tar.gz

# Add to PATH
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
echo 'export GOPATH=$HOME/go' >> ~/.bashrc
source ~/.bashrc

# Verify installation
go version
```

#### 2. Install Java

```bash
# Ubuntu/Debian
sudo apt update
sudo apt install openjdk-21-jdk

# macOS (using Homebrew)
brew install openjdk@21

# Verify installation
java -version
```

#### 3. Install Maven

```bash
# Ubuntu/Debian
sudo apt install maven

# macOS
brew install maven

# Verify installation
mvn -version
```

#### 4. Install Protocol Buffers Compiler

```bash
# Download protoc
wget https://github.com/protocolbuffers/protobuf/releases/download/v25.0/protoc-25.0-linux-x86_64.zip
sudo unzip protoc-25.0-linux-x86_64.zip -d /usr/local

# Install Go plugins
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Verify installation
protoc --version
```

#### 5. Install Docker

```bash
# Install Docker Engine
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh

# Add user to docker group
sudo usermod -aG docker $USER

# Install Docker Compose
sudo curl -L "https://github.com/docker/compose/releases/download/v2.20.0/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose

# Verify installation
docker --version
docker-compose --version
```

### IDE Setup

#### VS Code (Recommended for Go)

Install extensions:
```bash
code --install-extension golang.go
code --install-extension ms-azuretools.vscode-docker
code --install-extension zxh404.vscode-proto3
code --install-extension humao.rest-client
```

**settings.json:**
```json
{
  "go.useLanguageServer": true,
  "go.lintTool": "golangci-lint",
  "go.formatTool": "goimports",
  "go.testFlags": ["-v"],
  "go.coverOnSave": true,
  "editor.formatOnSave": true,
  "editor.codeActionsOnSave": {
    "source.organizeImports": true
  }
}
```

#### IntelliJ IDEA (Recommended for Java)

Install plugins:
- Spring Boot
- Docker
- Protocol Buffers
- Lombok

---

## 📁 Project Structure

```
Real-Time-Financial-News-Market-Impact-Analysis-system/
├── docs/                           # Documentation
│   ├── api/
│   │   ├── grpc-apis.md
│   │   └── rest-apis.md
│   ├── architecture.md
│   ├── deployment.md
│   └── development.md
├── services/                       # Microservices
│   ├── news-ingestion/            # Go service
│   │   ├── cmd/
│   │   │   └── server/
│   │   │       └── main.go
│   │   ├── internal/
│   │   │   ├── client/
│   │   │   ├── database/
│   │   │   ├── handler/
│   │   │   ├── model/
│   │   │   ├── repository/
│   │   │   └── service/
│   │   ├── proto/
│   │   ├── config/
│   │   ├── Dockerfile
│   │   ├── go.mod
│   │   └── README.md
│   ├── nlp-processing/            # Go service
│   │   └── [similar structure]
│   ├── MarketImpact/              # Java service
│   │   ├── src/
│   │   │   ├── main/
│   │   │   │   ├── java/
│   │   │   │   ├── proto/
│   │   │   │   └── resources/
│   │   │   └── test/
│   │   ├── pom.xml
│   │   ├── Dockerfile
│   │   └── README.md
│   └── AlertSignal/               # Java service
│       └── [similar structure]
├── k8s/                           # Kubernetes manifests
│   ├── production/
│   ├── staging/
│   └── dev/
├── scripts/                       # Utility scripts
│   ├── setup.sh
│   ├── test.sh
│   └── deploy.sh
├── docker-compose.yml
└── README.md
```

### Service Architecture Pattern

```mermaid
graph TB
    subgraph "Go Service Structure"
        MAIN[main.go<br/>Entry Point]
        CONFIG[config/<br/>Configuration]
        HANDLER[handler/<br/>HTTP & gRPC]
        SERVICE[service/<br/>Business Logic]
        REPO[repository/<br/>Data Access]
        MODEL[model/<br/>Data Models]
        CLIENT[client/<br/>External Clients]
        DB[(Database)]
    end

    MAIN --> CONFIG
    MAIN --> HANDLER
    HANDLER --> SERVICE
    SERVICE --> REPO
    SERVICE --> CLIENT
    REPO --> MODEL
    REPO --> DB

    style MAIN fill:#00ADD8,stroke:#00758F
    style SERVICE fill:#4CAF50,stroke:#2E7D32
    style REPO fill:#FF9800,stroke:#E65100
```

```mermaid
graph TB
    subgraph "Java Service Structure"
        MAIN_JAVA[Application.java<br/>Spring Boot Entry]
        CONFIG_JAVA[config/<br/>@Configuration]
        CONTROLLER[controller/<br/>@RestController]
        SERVICE_JAVA[service/<br/>@Service]
        REPO_JAVA[repository/<br/>@Repository]
        ENTITY[entity/<br/>@Entity]
        DTO[dto/<br/>Data Transfer]
        GRPC_JAVA[grpc/<br/>gRPC Services]
        DB_JAVA[(Database)]
    end

    MAIN_JAVA --> CONFIG_JAVA
    MAIN_JAVA --> CONTROLLER
    MAIN_JAVA --> GRPC_JAVA
    CONTROLLER --> SERVICE_JAVA
    GRPC_JAVA --> SERVICE_JAVA
    SERVICE_JAVA --> REPO_JAVA
    REPO_JAVA --> ENTITY
    SERVICE_JAVA --> DTO
    REPO_JAVA --> DB_JAVA

    style MAIN_JAVA fill:#ED8B00,stroke:#B86A00
    style SERVICE_JAVA fill:#6DB33F,stroke:#4E7B2B
    style REPO_JAVA fill:#FF9800,stroke:#E65100
```

---

## 📝 Coding Standards

### Go Code Standards

#### Formatting

```bash
# Format code
go fmt ./...

# Organize imports
goimports -w .

# Run linter
golangci-lint run
```

#### Naming Conventions

```go
// ✅ Good
type UserService interface {
    GetUser(ctx context.Context, id string) (*User, error)
    CreateUser(ctx context.Context, user *User) error
}

// ❌ Bad
type user_service interface {
    get_user(ctx context.Context, id string) (*User, error)
}
```

#### Error Handling

```go
// ✅ Good
func ProcessArticle(ctx context.Context, article *Article) error {
    if article == nil {
        return fmt.Errorf("article cannot be nil")
    }
    
    if err := validate(article); err != nil {
        return fmt.Errorf("validation failed: %w", err)
    }
    
    return nil
}

// ❌ Bad
func ProcessArticle(article *Article) error {
    validate(article) // Ignoring error
    return nil
}
```

#### Context Usage

```go
// ✅ Good
func FetchData(ctx context.Context, url string) ([]byte, error) {
    req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
    if err != nil {
        return nil, err
    }
    
    resp, err := client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    return io.ReadAll(resp.Body)
}
```

#### Logging

```go
import "github.com/sirupsen/logrus"

log := logrus.WithFields(logrus.Fields{
    "service": "news-ingestion",
    "article_id": articleID,
})

log.Info("Processing article")
log.WithError(err).Error("Failed to process article")
```

### Java Code Standards

#### Spring Boot Best Practices

```java
// ✅ Good
@Service
@Slf4j
public class MarketPredictionService {
    
    private final MarketPredictionRepository repository;
    private final NLPClient nlpClient;
    
    @Autowired
    public MarketPredictionService(
            MarketPredictionRepository repository,
            NLPClient nlpClient) {
        this.repository = repository;
        this.nlpClient = nlpClient;
    }
    
    public MarketPrediction createPrediction(CreatePredictionRequest request) {
        log.info("Creating prediction for symbol: {}", request.getSymbol());
        
        try {
            // Business logic
            return prediction;
        } catch (Exception e) {
            log.error("Failed to create prediction", e);
            throw new PredictionException("Failed to create prediction", e);
        }
    }
}
```

#### Exception Handling

```java
// ✅ Good
@ControllerAdvice
public class GlobalExceptionHandler {
    
    @ExceptionHandler(ResourceNotFoundException.class)
    public ResponseEntity<ErrorResponse> handleNotFound(ResourceNotFoundException ex) {
        ErrorResponse error = new ErrorResponse(
            "NOT_FOUND",
            ex.getMessage(),
            LocalDateTime.now()
        );
        return ResponseEntity.status(HttpStatus.NOT_FOUND).body(error);
    }
}
```

#### Lombok Usage

```java
// ✅ Good
@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
@Entity
@Table(name = "market_predictions")
public class MarketPrediction {
    
    @Id
    @GeneratedValue(generator = "UUID")
    @GenericGenerator(name = "UUID", strategy = "org.hibernate.id.UUIDGenerator")
    private UUID id;
    
    @Column(nullable = false)
    private String symbol;
    
    @Column(name = "predicted_change_percent")
    private BigDecimal predictedChangePercent;
    
    private String direction;
    private Double confidence;
    
    @CreationTimestamp
    private LocalDateTime predictionTimestamp;
}
```

---

## 🔀 Git Workflow

### Branch Strategy

```mermaid
gitGraph
    commit id: "Initial commit"
    branch develop
    checkout develop
    commit id: "Feature base"
    
    branch feature/news-api
    checkout feature/news-api
    commit id: "Add NewsAPI client"
    commit id: "Add tests"
    
    checkout develop
    merge feature/news-api
    
    branch release/v1.0
    checkout release/v1.0
    commit id: "Prepare release"
    
    checkout main
    merge release/v1.0 tag: "v1.0.0"
    
    checkout develop
    merge release/v1.0
```

### Commit Message Convention

Follow **Conventional Commits**:

```bash
# Format
<type>(<scope>): <subject>

<body>

<footer>
```

**Types:**
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation
- `style`: Code style (formatting)
- `refactor`: Code refactoring
- `test`: Tests
- `chore`: Maintenance

**Examples:**
```bash
feat(news-ingestion): add NewsAPI client integration

Implement NewsAPI client with rate limiting and error handling.
Includes configuration for API key and source filtering.

Closes #123

---

fix(nlp-processing): resolve sentiment calculation edge case

Handle empty article content gracefully without throwing exception.

Fixes #456

---

docs(api): update gRPC API documentation

Add examples for StreamArticles RPC with filters.
```

### Pull Request Template

```markdown
## Description
Brief description of changes

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation update

## Testing
- [ ] Unit tests added/updated
- [ ] Integration tests added/updated
- [ ] Manual testing completed

## Checklist
- [ ] Code follows project style guidelines
- [ ] Self-review completed
- [ ] Comments added for complex code
- [ ] Documentation updated
- [ ] No new warnings generated
- [ ] Tests pass locally
```

---

## 🧪 Testing

### Testing Pyramid

```mermaid
graph TB
    subgraph "Testing Levels"
        E2E[End-to-End Tests<br/>10%]
        INT[Integration Tests<br/>20%]
        UNIT[Unit Tests<br/>70%]
    end

    E2E --> INT
    INT --> UNIT

    style E2E fill:#F44336,stroke:#C62828,color:#fff
    style INT fill:#FF9800,stroke:#E65100,color:#fff
    style UNIT fill:#4CAF50,stroke:#2E7D32,color:#fff
```

### Go Testing

#### Unit Tests

```go
// service_test.go
package service

import (
    "context"
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
)

type MockRepository struct {
    mock.Mock
}

func (m *MockRepository) Save(ctx context.Context, article *Article) error {
    args := m.Called(ctx, article)
    return args.Error(0)
}

func TestIngestionService_ProcessArticle(t *testing.T) {
    // Arrange
    mockRepo := new(MockRepository)
    service := NewIngestionService(mockRepo)
    
    article := &Article{
        Title:   "Test Article",
        Content: "Test content",
    }
    
    mockRepo.On("Save", mock.Anything, article).Return(nil)
    
    // Act
    err := service.ProcessArticle(context.Background(), article)
    
    // Assert
    assert.NoError(t, err)
    mockRepo.AssertExpectations(t)
}
```

#### Table-Driven Tests

```go
func TestValidateArticle(t *testing.T) {
    tests := []struct {
        name    string
        article *Article
        wantErr bool
    }{
        {
            name: "valid article",
            article: &Article{
                Title:   "Valid Title",
                Content: "Valid content with sufficient length",
            },
            wantErr: false,
        },
        {
            name: "empty title",
            article: &Article{
                Title:   "",
                Content: "Valid content",
            },
            wantErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ValidateArticle(tt.article)
            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

#### Run Tests

```bash
# Run all tests
go test ./...

# Run with coverage
go test ./... -cover

# Generate coverage report
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out

# Run specific test
go test -run TestIngestionService_ProcessArticle

# Run with race detector
go test -race ./...

# Verbose output
go test -v ./...
```

### Java Testing

#### Unit Tests with JUnit 5

```java
@ExtendWith(MockitoExtension.class)
class MarketPredictionServiceTest {
    
    @Mock
    private MarketPredictionRepository repository;
    
    @Mock
    private NLPClient nlpClient;
    
    @InjectMocks
    private MarketPredictionService service;
    
    @Test
    void createPrediction_ValidRequest_ReturnsP prediction() {
        // Arrange
        String symbol = "AAPL";
        SentimentData sentiment = new SentimentData(0.85, 0.92);
        
        when(nlpClient.getSentimentTrends(symbol, 24))
            .thenReturn(sentiment);
        
        // Act
        MarketPrediction result = service.createPrediction(symbol);
        
        // Assert
        assertNotNull(result);
        assertEquals(symbol, result.getSymbol());
        assertEquals("UP", result.getDirection());
        verify(repository).save(any(MarketPrediction.class));
    }
}
```

#### Integration Tests

```java
@SpringBootTest(webEnvironment = WebEnvironment.RANDOM_PORT)
@TestPropertySource(locations = "classpath:application-test.properties")
class MarketImpactControllerIntegrationTest {
    
    @Autowired
    private TestRestTemplate restTemplate;
    
    @Test
    void getPredictions_ReturnsListOfPredictions() {
        // Act
        ResponseEntity<List<MarketPrediction>> response = 
            restTemplate.exchange(
                "/api/v1/market-predictions",
                HttpMethod.GET,
                null,
                new ParameterizedTypeReference<>() {}
            );
        
        // Assert
        assertEquals(HttpStatus.OK, response.getStatusCode());
        assertNotNull(response.getBody());
    }
}
```

#### Run Tests

```bash
# Run all tests
./mvnw test

# Run specific test class
./mvnw test -Dtest=MarketPredictionServiceTest

# Run with coverage
./mvnw clean verify

# Generate coverage report
./mvnw jacoco:report
```

---

## 🐛 Debugging

### Go Debugging with Delve

#### Install Delve

```bash
go install github.com/go-delve/delve/cmd/dlv@latest
```

#### Debug Application

```bash
# Debug main package
dlv debug main.go

# Set breakpoint
(dlv) break main.main
(dlv) break service.ProcessArticle

# Run
(dlv) continue

# Step through
(dlv) next
(dlv) step

# Inspect variables
(dlv) print articleID
(dlv) locals

# Exit
(dlv) exit
```

#### VS Code Launch Configuration

`.vscode/launch.json`:
```json
{
    "version": "0.2.0",
    "configurations": [
        {
            "name": "Debug News Ingestion",
            "type": "go",
            "request": "launch",
            "mode": "debug",
            "program": "${workspaceFolder}/services/news-ingestion/main.go",
            "env": {
                "POSTGRES_HOST": "localhost",
                "POSTGRES_PASSWORD": "postgres123"
            }
        }
    ]
}
```

### Java Debugging

#### IntelliJ IDEA

1. Set breakpoints by clicking left gutter
2. Run > Debug 'Application'
3. Use debug panel to step through code

#### Remote Debugging

```bash
# Start Java application with debug port
java -agentlib:jdwp=transport=dt_socket,server=y,suspend=n,address=*:5005 \
  -jar target/MarketImpact-0.0.1-SNAPSHOT.jar
```

IntelliJ Configuration:
```
Run > Edit Configurations > Remote JVM Debug
Host: localhost
Port: 5005
```

---

## 🔧 Code Generation

### Protocol Buffers

#### Generate Go Code

```bash
cd services/news-ingestion
protoc --go_out=. --go_opt=paths=source_relative \
       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
       proto/NewsService.proto
```

#### Generate Java Code

```bash
cd services/MarketImpact
mvn clean compile
# Maven automatically generates from src/main/proto/
```

### Database Models

#### GORM Generate (Go)

```go
// Use gen tool
go install gorm.io/gen/tools/gentool@latest

// Generate models
gentool -dsn "postgres://user:pass@localhost/dbname?sslmode=disable" \
        -outPath "./internal/model"
```

---

## 💾 Database Migrations

### Go Migrations (golang-migrate)

```bash
# Install
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Create migration
migrate create -ext sql -dir migrations -seq create_articles_table

# Apply migrations
migrate -path migrations -database "postgres://localhost/dbname?sslmode=disable" up

# Rollback
migrate -path migrations -database "postgres://localhost/dbname?sslmode=disable" down 1
```

### Java Migrations (Flyway)

Migrations in `src/main/resources/db/migration/`:
```sql
-- V1__Create_market_predictions.sql
CREATE TABLE market_predictions (
    id UUID PRIMARY KEY,
    symbol VARCHAR(10) NOT NULL,
    predicted_change_percent DECIMAL(10, 4),
    direction VARCHAR(10),
    confidence DECIMAL(10, 4),
    prediction_timestamp TIMESTAMP
);
```

Flyway runs automatically on Spring Boot startup.

---

## ⚡ Performance Optimization

### Go Performance

#### Profiling

```bash
# CPU profiling
go test -cpuprofile=cpu.prof -bench=.
go tool pprof cpu.prof

# Memory profiling
go test -memprofile=mem.prof -bench=.
go tool pprof mem.prof

# Web interface
go tool pprof -http=:8080 cpu.prof
```

#### Benchmarking

```go
func BenchmarkProcessArticle(b *testing.B) {
    service := NewIngestionService(mockRepo)
    article := &Article{Title: "Benchmark", Content: "Content"}
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        service.ProcessArticle(context.Background(), article)
    }
}
```

### Java Performance

#### JVM Options

```bash
# Heap size
-Xms512m -Xmx2g

# GC tuning
-XX:+UseG1GC -XX:MaxGCPauseMillis=200

# JIT optimization
-XX:+TieredCompilation -XX:TieredStopAtLevel=1
```

---

## 🤝 Contributing

### Contribution Workflow

```mermaid
graph LR
    A[Fork Repository] --> B[Create Branch]
    B --> C[Make Changes]
    C --> D[Write Tests]
    D --> E[Run Tests]
    E --> F[Commit Changes]
    F --> G[Push to Fork]
    G --> H[Create PR]
    H --> I[Code Review]
    I --> J[Merge]

    style A fill:#4CAF50,stroke:#2E7D32
    style J fill:#2196F3,stroke:#1565C0
```

### Code Review Checklist

- [ ] Code follows project conventions
- [ ] Tests added and passing
- [ ] Documentation updated
- [ ] No breaking changes (or documented)
- [ ] Performance impact considered
- [ ] Security implications reviewed
- [ ] Error handling adequate
- [ ] Logging appropriate

---

<div align="center">

**Happy Coding! 🚀**

![Go](https://img.shields.io/badge/Go-00ADD8?style=flat&logo=go&logoColor=white)
![Java](https://img.shields.io/badge/Java-ED8B00?style=flat&logo=openjdk&logoColor=white)
![Docker](https://img.shields.io/badge/Docker-2496ED?style=flat&logo=docker&logoColor=white)

[Back to Top](#-development-guide)

</div>