# User Management Service

A user management API with idempotent operations powered by Temporal workflows.

## Overview

This service exposes REST APIs for user management:
- **Create User** - Register new users
- **Get User** - Retrieve user details  
- **Flag User** - Apply flags to users (fraud, suspended, etc.)
- **Lift Flag** - Remove flags from users

All mutating operations are **idempotent** (via `referenceID` in payload) and execute as **Temporal workflows** in the background.

## Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                                                                             │
│  ┌──────────┐      ┌─────────────────┐      ┌──────────────────────────┐   │
│  │          │      │                 │      │     Temporal Server      │   │
│  │  Client  │─────▶│   API Server    │─────▶│                          │   │
│  │          │      │                 │      │  ┌────────────────────┐  │   │
│  └──────────┘      └─────────────────┘      │  │  Workflow Queue    │  │   │
│       │                    │                │  └─────────┬──────────┘  │   │
│       │                    │                └────────────┼─────────────┘   │
│       │                    │                             │                 │
│       │            ┌───────▼───────┐                     │                 │
│       │            │     Redis     │            ┌────────▼─────────┐       │
│       │            │ (Idempotency) │            │      Worker      │       │
│       │            └───────────────┘            │  (same binary)   │       │
│       │                                         └────────┬─────────┘       │
│       │                                                  │                 │
│       │         poll status                     ┌────────▼─────────┐       │
│       └────────────────────────────────────────▶│      MySQL       │       │
│                                                 └──────────────────┘       │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

<details>
<summary>Mermaid version</summary>

```mermaid
flowchart TB
    subgraph Service["User Management Service"]
        Client([Client])
        API[API Server]
        Redis[(Redis<br/>Idempotency)]
        
        subgraph Temporal["Temporal Server"]
            Queue[Workflow Queue]
        end
        
        Worker[Worker<br/>same binary]
        MySQL[(MySQL)]
    end
    
    Client -->|HTTP Request| API
    API -->|Start Workflow| Queue
    API -->|Check/Store referenceID| Redis
    Queue -->|Execute| Worker
    Worker -->|Read/Write| MySQL
    Client -.->|Poll Status| API
```

</details>

## API Flow

```
  Client                    API Server                 Temporal                  Worker
    │                           │                         │                        │
    │  POST /users              │                         │                        │
    │  {referenceID: "abc"}     │                         │                        │
    ├──────────────────────────▶│                         │                        │
    │                           │                         │                        │
    │                           │  Start Workflow         │                        │
    │                           │  (ID = referenceID)     │                        │
    │                           ├────────────────────────▶│                        │
    │                           │                         │                        │
    │  202 Accepted             │                         │   Execute              │
    │  {workflowID: "abc"}      │                         │   Workflow             │
    │◀──────────────────────────┤                         ├───────────────────────▶│
    │                           │                         │                        │
    │                           │                         │                        │
    │  GET /workflows/abc       │                         │                        │
    ├──────────────────────────▶│  Query Status           │                        │
    │                           ├────────────────────────▶│                        │
    │  {status: "completed"}    │                         │                        │
    │◀──────────────────────────┤                         │                        │
    │                           │                         │                        │
```

<details>
<summary>Mermaid version</summary>

```mermaid
sequenceDiagram
    participant C as Client
    participant A as API Server
    participant T as Temporal
    participant W as Worker

    C->>A: POST /users<br/>{referenceID: "abc"}
    A->>T: Start Workflow<br/>(ID = referenceID)
    A-->>C: 202 Accepted<br/>{workflowID: "abc"}
    
    T->>W: Execute Workflow
    W->>W: Run Activities
    W-->>T: Complete
    
    C->>A: GET /workflows/abc
    A->>T: Query Status
    T-->>A: Status + Result
    A-->>C: {status: "completed"}
```

</details>

## API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/api/v1/users` | Create user (workflow) |
| `GET` | `/api/v1/users/:id` | Get user |
| `POST` | `/api/v1/users/:id/flags` | Add flag (workflow) |
| `DELETE` | `/api/v1/users/:id/flags/:type` | Remove flag (workflow) |
| `GET` | `/api/v1/workflows/:id` | Poll workflow status |

## Quick Start

```bash
# Start all services
make up

# Run migrations
make migrate-up

# Run the application
make run
```

**Access:** http://localhost:8080

## License

MIT
