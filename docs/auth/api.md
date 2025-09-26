# Auth API Documentation

This document describes the authentication endpoints for the Be-Ambis-Solving API.

## Base URL
All endpoints are prefixed with `/api`.

## Endpoints

### Login
Authenticate a user and return a JWT token.

- **Method**: `POST`
- **Path**: `/login`
- **Content-Type**: `application/json`

#### Request Body
```json
{
  "email": "string",
  "password": "string"
}
```

#### Response (200 OK)
```json
{
  "token": "string",
  "userId": "string"
}
```

#### Error Responses
- **400 Bad Request**: Invalid request body
  ```json
  {
    "error": "invalid body"
  }
  ```
- **401 Unauthorized**: Invalid email or password
  ```json
  {
    "error": "invalid email or password"
  }
  ```

#### Example
```bash
curl -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
  -d '{"email": "user@example.com", "password": "password123"}'
```

### Register
Create a new user account. This endpoint is only enabled if `ENABLE_REGISTER=true` in the configuration.

- **Method**: `POST`
- **Path**: `/register`
- **Content-Type**: `application/json`

#### Request Body
```json
{
  "name": "string",
  "email": "string",
  "password": "string"
}
```

#### Response (201 Created)
```json
{
  "id": "string",
  "email": "string",
  "name": "string"
}
```

#### Error Responses
- **403 Forbidden**: Registration disabled
  ```json
  {
    "error": "registration disabled"
  }
  ```
- **400 Bad Request**: Invalid request body or user creation error
  ```json
  {
    "error": "invalid body"
    // or specific error message from user creation
  }
  ```

#### Example
```bash
curl -X POST http://localhost:8080/api/register \
  -H "Content-Type: application/json" \
  -d '{"name": "John Doe", "email": "john@example.com", "password": "password123"}'
```

## Notes
- Both endpoints are public and do not require authentication.
- The login endpoint returns a JWT token that should be used for subsequent authenticated requests.
- Registration is disabled by default for security reasons and can be enabled via configuration.