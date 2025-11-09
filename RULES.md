# Rival App - Development Rules

## 🏗️ Architecture Rules

### Project Structure
```
internal/
├── auth/           # Authentication module
├── user/           # User profile & photo management
├── wallet/         # Wallet & transactions
├── coin/           # Coin system
├── notification/   # Notifications
├── reward/         # Rewards system
├── activity/       # User activities
├── global/         # Global configurations
└── interface/      # Interfaces for all modules

uploads/
└── profile_photos/ # Local photo storage
```

### Layer Separation
- **Handler** - Only API endpoints, request/response handling
- **Service** - Business logic only
- **Repository** - Database operations only
- **Utils** - Utility functions only

## 🔐 Authentication Rules

### JWT Token
- **Expiry**: 24 hours
- **Algorithm**: HS256
- **Claims**: user_id, email, exp, iat
- **Storage**: Hashed tokens in database for logout tracking

### Password Security
- **Hashing**: bcrypt
- **Validation**: Min 8 chars, 1 uppercase, 1 lowercase, 1 number
- **Reset**: OTP-based (5 min expiry)

### OTP System
- **Test OTP**: 123456 (always valid)
- **Generated OTP**: 6-digit random
- **Expiry**: 5 minutes
- **Storage**: In-memory map

## 👤 User Management Rules

### Profile Management
- **Name**: 2-100 characters, alphanumeric + spaces
- **Phone**: 10-15 digits, international format
- **DOB**: YYYY-MM-DD format only
- **Email**: Unique, validated format

### Photo Upload System
- **Formats**: JPG, JPEG, PNG, GIF only
- **Upload**: Base64 encoded strings
- **Storage**: Local directory `./uploads/profile_photos/`
- **Naming**: `userID_timestamp.ext` format
- **Size**: No limit (add if needed)
- **Security**: Path traversal protection

### Verification Rules
- **Email**: OTP-based verification
- **Phone**: OTP-based verification
- **Test OTP**: 123456 for all verifications

## 💰 Coin System Rules

### Coin Operations
- **Earn**: Add coins with reason tracking
- **Spend**: Deduct coins with validation
- **Balance**: Real-time coin balance
- **Packages**: Predefined coin packages with bonus

### Transaction Tracking
- All coin operations logged
- User ID, amount, type, reason stored
- Immutable transaction history

## 💳 Wallet Rules

### Wallet Operations
- **Add Money**: Credit wallet balance
- **Deduct Money**: Debit with validation
- **Transfer**: P2P money transfer with atomic transactions
- **Balance**: Separate from coins
- **Currency**: INR, USD, EUR supported

### Transaction History
- All wallet operations tracked
- Title, description, amount, type stored
- Paginated history API (20 per page)
- Icons for transaction types

### Validation Rules
- **Amount**: Positive, max 1,000,000
- **User ID**: Positive integer
- **Title**: 1-100 chars, alphanumeric + basic symbols
- **Description**: Max 500 characters
- **Page**: 1-1000 range

## 🔒 Security Rules

### API Security
- All endpoints require proper validation
- Sensitive operations need authentication
- Input sanitization mandatory
- File upload security (extension validation)

### Database Security
- Foreign key constraints enforced
- Cascade deletes for user data
- Indexed queries for performance
- Transaction rollback on failures

### File Security
- Only allowed file extensions
- Unique filename generation
- Path traversal prevention
- Automatic cleanup on errors

## 📝 Code Standards

### Naming Conventions
- **Files**: snake_case.go
- **Functions**: PascalCase for exported, camelCase for private
- **Variables**: camelCase
- **Constants**: UPPER_SNAKE_CASE
- **Photo Files**: userID_timestamp.ext

### Error Handling
- Always return meaningful error messages
- Use custom error types where needed
- Log errors for debugging
- Rollback operations on failures

### API Response Format
```json
{
  "data": {},
  "message": "Success message",
  "error": null
}
```

## 🚀 Development Rules

### Database Changes
1. Create migration file in `sql/schema/`
2. Add queries in `sql/query/`
3. Run `sqlc generate`
4. Update repository methods

### New Feature Addition
1. Define interface in `internal/interface/`
2. Implement repository layer
3. Implement service layer
4. Create handlers
5. Add tests

### File Upload Rules
1. Validate file extension first
2. Decode Base64 data
3. Create directory if needed
4. Generate unique filename
5. Save file to local storage
6. Update database record
7. Rollback file on DB error

### Testing Rules
- Test OTP: 123456
- Test credentials: Any valid email/password
- Mock external services
- Unit tests for business logic

## 📊 API Endpoints

### Authentication
- `POST /auth/signup` - User registration
- `POST /auth/login` - User login
- `POST /auth/verify` - Token verification
- `POST /auth/logout` - User logout
- `POST /auth/forgot-password` - Send reset OTP
- `POST /auth/reset-password` - Reset with OTP
- `POST /auth/send-otp` - Send OTP
- `POST /auth/verify-otp` - Verify OTP

### User Management
- `GET /user/profile/:userID` - Get user profile
- `PUT /user/profile` - Update profile
- `POST /user/upload-photo` - Upload profile photo
- `GET /user/photo/:fileName` - Get uploaded photo
- `DELETE /user/photo/:userID` - Delete profile photo
- `GET /user/photos/:userID` - List user photos
- `POST /user/verify-email` - Verify email with OTP
- `POST /user/verify-phone` - Verify phone with OTP
- `GET /user/dashboard/:userID` - User dashboard
- `GET /user/test` - Test user system

### Wallet
- `GET /wallet/balance/:userID` - Get wallet balance
- `POST /wallet/create` - Create new wallet
- `POST /wallet/add` - Add money
- `POST /wallet/deduct` - Deduct money
- `POST /wallet/transfer` - Transfer money P2P
- `GET /wallet/history/:userID/:page` - Transaction history
- `GET /wallet/stats/:userID` - Wallet statistics
- `GET /wallet/summary/:userID` - Complete wallet summary
- `GET /wallet/test` - Test wallet system

### Coins
- `GET /coins/balance/:userID` - Get coin balance
- `POST /coins/earn` - Earn coins
- `POST /coins/spend` - Spend coins
- `GET /coins/packages` - Get coin packages
- `POST /coins/purchase` - Purchase coins

## ⚠️ Important Notes

- Never commit sensitive data
- Always validate user input
- Use proper HTTP status codes
- Maintain backward compatibility
- Document API changes
- Test before deployment
- Clean up files on errors
- Use atomic transactions for critical operations

## 🔧 Environment Setup

### Required Tools
- Go 1.21+
- PostgreSQL
- Encore CLI
- SQLC

### Environment Variables
```
GOOSE_DRIVER=postgres
GOOSE_DBSTRING=postgres://user:pass@host:port/db
GOOSE_MIGRATION_DIR=./sql/schema
```

### Directory Structure
```
uploads/
└── profile_photos/     # User profile photos
    └── userID_timestamp.ext
```

## 📁 File Management Rules

### Upload Process
1. **Validation**: Check file extension (.jpg, .jpeg, .png, .gif)
2. **Decoding**: Convert Base64 to binary data
3. **Storage**: Save to `./uploads/profile_photos/`
4. **Naming**: Format `userID_timestamp.extension`
5. **Database**: Update user profile_photo field
6. **Cleanup**: Remove file if DB update fails

### File Serving
- **Security**: Only serve files from allowed directories
- **MIME Types**: Proper content-type headers
- **Base64**: Return images as Base64 encoded strings
- **Access Control**: User-specific file access

### File Deletion
- **Database**: Remove photo path from user record
- **Filesystem**: Delete actual file from storage
- **Validation**: Ensure user owns the photo

---
**Last Updated**: November 2024
**Version**: 2.0.0
