# RIVAL Development Rules & Architecture

## Project Overview
RIVAL is a reward coin system for restaurants where users buy coins and get discounts. Built with Go backend, PostgreSQL, Redis, TigerBeetle for transactions, MinIO for storage, MailHog for emails, and Firebase for social auth.

## Core Architecture Principles

### 1. Clean Architecture Layers
```
Handler (gRPC) → Service (Business Logic) → Repository (Database)
```

**NEVER:**
- Pass protobuf types to service layer
- Return protobuf types from repository
- Mix business logic in handlers

**ALWAYS:**
- Convert gRPC requests to service parameters in handler
- Service returns gRPC response types
- Repository returns sqlc generated types
- Service converts between sqlc and protobuf types

### 2. Directory Structure
```
internal/
├── auth/
│   ├── handler/     # gRPC handlers
│   ├── service/     # Business logic
│   ├── repo/        # Database operations
│   └── util/        # Utilities (JWT, Email, Firebase)
├── users/
├── merchants/
└── payments/
```

### 3. Database & Code Generation

**SQL Schema:**
- Location: `sql/schema/001_initial_schema.sql`
- Use goose for migrations: `-- +goose Up` / `-- +goose Down`
- Always include proper indexes

**SQL Queries:**
- Location: `sql/queries/{module}.sql`
- Use sqlc annotations: `-- name: FunctionName :one/:many/:exec`
- Generated code goes to: `gen/sql/`

**Protobuf:**
- Schema: `proto/schema/schema.proto`
- APIs: `proto/api/{service}.proto`
- Generated code goes to: `gen/proto/`

**Generation Commands:**
```bash
make proto-gen    # Generate protobuf
make sqlc-gen     # Generate SQL code
make gen-all      # Generate everything
```

### 4. Service Implementation Pattern

**Handler Example:**
```go
func (h *AuthHandler) Signup(ctx context.Context, req *authpb.SignupRequest) (*authpb.SignupResponse, error) {
    // 1. Validate input
    if req.Email == "" {
        return &authpb.SignupResponse{Message: "Email required"}, nil
    }

    // 2. Convert to service parameters
    params := service.SignupParams{
        Email: req.Email,
        Password: req.Password,
    }

    // 3. Call service
    return h.service.Signup(ctx, params)
}
```

**Service Example:**
```go
func (s *authService) Signup(ctx context.Context, params SignupParams) (*authpb.SignupResponse, error) {
    // 1. Business logic
    hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(params.Password), bcrypt.DefaultCost)

    // 2. Convert to repository parameters
    createParams := schema.CreateUserParams{
        Email: params.Email,
        PasswordHash: pgtype.Text{String: string(hashedPassword), Valid: true},
    }

    // 3. Call repository
    user, err := s.repo.CreateUser(ctx, createParams)

    // 4. Convert to protobuf and return
    return &authpb.SignupResponse{Message: "Success"}, nil
}
```

**Repository Example:**
```go
func (r *authRepository) CreateUser(ctx context.Context, params schema.CreateUserParams) (schema.User, error) {
    // 1. Database operation
    user, err := r.queries.CreateUser(ctx, params)

    // 2. TigerBeetle account creation
    accountID := uuidToUint128(user.ID.String())
    account := types.Account{ID: accountID, Ledger: 1, Code: 1}
    r.tb.CreateAccounts([]types.Account{account})

    return user, err
}
```

### 5. Database Integration

**PostgreSQL (Primary Database):**
- Use pgx/v5 with pgxpool
- All user data, transactions, sessions
- Connection: `connection.GetPgConnection()`

**Redis (Cache & OTP):**
- OTP storage with expiry
- Session caching
- Connection: `connection.GetRedisClient()`

**TigerBeetle (Financial Transactions):**
- Account creation for each user (same UUID as DB)
- All coin transactions
- Connection: `connection.NewTbClient()`

**MinIO (File Storage):**
- Profile pictures, documents
- Use signed URLs for uploads
- Connection: `connection.NewMinioClient()`

### 6. Type Handling

**PostgreSQL Types:**
```go
// String fields
pgtype.Text{String: value, Valid: value != ""}

// Timestamps
pgtype.Timestamp{Time: time.Now(), Valid: true}

// UUIDs
pgtype.UUID{} // Use .Scan() method

// Numeric
pgtype.Numeric{} // Use .Value() method
```

**Protobuf Conversion:**
```go
func convertToProtoUser(user schema.User) *schemapb.User {
    userID, _ := user.ID.Value()
    return &schemapb.User{
        Id: userID.(string),
        Email: user.Email,
        CreatedAt: user.CreatedAt.Time.Unix(),
    }
}
```

### 7. Error Handling

**Service Layer:**
- Return business logic errors
- Use fmt.Errorf for context
- Don't expose internal errors to client

**Handler Layer:**
- Validate input and return proper gRPC errors
- Convert service errors to appropriate responses

### 8. Configuration

**Config Structure:**
```go
type Config struct {
    Database DatabaseConfig `yaml:"database"`
    Redis    RedisConfig    `yaml:"redis"`
    Firebase FirebaseConfig `yaml:"firebase"`
    MailHog  MailHogConfig  `yaml:"mail"`
    // etc...
}
```

**Usage:**
```go
cfg := config.GetConfig()
```

### 9. External Services

**Email (MailHog):**
```go
emailService := util.NewEmailService()
emailService.SendOTP(email, otp)
```

**Firebase Auth:**
```go
firebaseService, _ := util.NewFirebaseService(credentialsPath)
user, _ := firebaseService.VerifyToken(ctx, token)
```

**JWT:**
```go
jwtUtil := util.NewJWTUtil(secret, accessTTL, refreshTTL)
accessToken, refreshToken, _ := jwtUtil.GenerateTokens(user)
```

### 10. Testing & Building

**Build Commands:**
```bash
go build ./internal/auth/...     # Build specific module
go build ./...                   # Build everything
```

**Module Structure:**
- Each module (auth, users, merchants) is self-contained
- Handler creates all dependencies internally
- No shared state between modules

### 11. Security Rules

**Passwords:**
- Always use bcrypt for hashing
- Minimum 8 characters (implement in validation)

**JWT Tokens:**
- Store hashed tokens in database
- Implement token revocation
- Use proper expiry times

**Database:**
- Use parameterized queries (sqlc handles this)
- Validate all inputs in handlers
- Use proper indexes for performance

### 12. Business Logic Rules

**User Creation:**
- Create PostgreSQL user record
- Create TigerBeetle account with same UUID
- Generate unique referral code
- Send welcome email

**Coin System:**
- 1 Coin = $1 USD
- Restaurant discount: 15%
- Grocery discount: 2%
- All transactions through TigerBeetle

**Referral System:**
- Unique referral codes per user
- Reward both referrer and referred
- Track in referral_rewards table

### 13. API Design

**Protobuf Naming:**
- Services: `{Module}Service`
- Methods: `{Action}{Resource}`
- Messages: `{Action}{Resource}Request/Response`

**Streaming APIs:**
- Use for real-time updates
- Pattern: `Stream{Resource}Updates`

### 14. Future Development

**Adding New Module:**
1. Create directory: `internal/{module}/`
2. Add subdirectories: `handler/`, `service/`, `repo/`, `util/`
3. Create protobuf: `proto/api/{module}.proto`
4. Add SQL queries: `sql/queries/{module}.sql`
5. Follow same architecture pattern

**Adding New Feature:**
1. Add to protobuf schema if needed
2. Add SQL queries
3. Implement repository methods
4. Implement service methods
5. Add handler methods
6. Generate code: `make gen-all`
7. Build and test

## Common Mistakes to Avoid

1. **Don't** pass protobuf types to service layer
2. **Don't** return protobuf types from repository
3. **Don't** mix database logic in service layer
4. **Don't** forget to handle PostgreSQL null types
5. **Don't** forget to create TigerBeetle accounts for users
6. **Don't** forget to validate inputs in handlers
7. **Don't** forget to regenerate code after schema changes
8. **Don't** forget to add proper error handling
9. **Don't** forget to use transactions for multi-step operations
10. **Don't** forget to test build after changes

## Module Dependencies

```
Handler → Service → Repository
   ↓        ↓         ↓
gRPC    Business   Database
Types    Logic     Operations
```

**External Dependencies:**
- PostgreSQL (via connection package)
- Redis (via connection package)
- TigerBeetle (via connection package)
- MinIO (via connection package)
- MailHog (via util package)
- Firebase (via util package)

alltrascation shuld bhi done with tigerbetle is main for transaction in financel related thing if money will be addded then in tigerbetle
tranfse then also


Follow these rules strictly for consistent, maintainable code!
