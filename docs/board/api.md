# Board API Documentation

This document describes the board management endpoints for the Be-Ambis-Solving API.

## Base URL
All endpoints are prefixed with `/api`.

## Authentication
All board endpoints require JWT authentication. Include the token in the Authorization header as `Bearer <token>`.

## Endpoints

### Create Board
Create a new board.

- **Method**: `POST`
- **Path**: `/boards`
- **Content-Type**: `application/json`

#### Request Body
```json
{
  "name": "string",
  "description": "string (optional)",
  "columns": [
    {
      "id": "string",
      "name": "string",
      "order": 0
    }
  ],
  "members": ["string"]
}
```

#### Response (201 Created)
Returns the created board.
```json
{
  "id": "string",
  "ownerId": "string",
  "name": "string",
  "description": "string (optional)",
  "members": ["string"],
  "columns": [
    {
      "id": "string",
      "name": "string",
      "order": 0
    }
  ],
  "isArchived": false,
  "createdAt": "string",
  "updatedAt": "string"
}
```

#### Error Responses
- **400 Bad Request**: Invalid request body or board creation error
  ```json
  {
    "error": "invalid body"
  }
  ```
- **401 Unauthorized**: Missing or invalid JWT token
  ```json
  {
    "error": "unauthorized"
  }
  ```

#### Example
```bash
curl -X POST http://localhost:8080/api/boards \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"name": "My Board", "columns": [{"id": "col1", "name": "To Do", "order": 0}]}'
```

### List Boards
Get all boards accessible to the authenticated user.

- **Method**: `GET`
- **Path**: `/boards`

#### Response (200 OK)
Array of boards.
```json
[
  {
    "id": "string",
    "ownerId": "string",
    "name": "string",
    "description": "string (optional)",
    "members": ["string"],
    "columns": [
      {
        "id": "string",
        "name": "string",
        "order": 0
      }
    ],
    "isArchived": false,
    "createdAt": "string",
    "updatedAt": "string"
  }
]
```

#### Error Responses
- **401 Unauthorized**: Missing or invalid JWT token
  ```json
  {
    "error": "unauthorized"
  }
  ```
- **500 Internal Server Error**: Server error
  ```json
  {
    "error": "internal server error"
  }
  ```

#### Example
```bash
curl -X GET http://localhost:8080/api/boards \
  -H "Authorization: Bearer <token>"
```

### Get Board
Get a specific board by ID.

- **Method**: `GET`
- **Path**: `/boards/:id`

#### Response (200 OK)
Board object.
```json
{
  "id": "string",
  "ownerId": "string",
  "name": "string",
  "description": "string (optional)",
  "members": ["string"],
  "columns": [
    {
      "id": "string",
      "name": "string",
      "order": 0
    }
  ],
  "isArchived": false,
  "createdAt": "string",
  "updatedAt": "string"
}
```

#### Error Responses
- **400 Bad Request**: Invalid board ID
  ```json
  {
    "error": "invalid id"
  }
  ```
- **401 Unauthorized**: Missing or invalid JWT token
  ```json
  {
    "error": "unauthorized"
  }
  ```
- **403 Forbidden**: User does not have access to the board
- **404 Not Found**: Board not found
  ```json
  {
    "error": "not found"
  }
  ```

#### Example
```bash
curl -X GET http://localhost:8080/api/boards/507f1f77bcf86cd799439011 \
  -H "Authorization: Bearer <token>"
```

### Update Board
Update an existing board.

- **Method**: `PATCH`
- **Path**: `/boards/:id`
- **Content-Type**: `application/json`

#### Request Body
All fields are optional.
```json
{
  "name": "string",
  "description": "string",
  "columns": [
    {
      "id": "string",
      "name": "string",
      "order": 0
    }
  ],
  "members": ["string"]
}
```

#### Response (204 No Content)

#### Error Responses
- **400 Bad Request**: Invalid request body or update error
  ```json
  {
    "error": "invalid body"
  }
  ```
- **401 Unauthorized**: Missing or invalid JWT token
  ```json
  {
    "error": "unauthorized"
  }
  ```
- **403 Forbidden**: User does not have access to the board

#### Example
```bash
curl -X PATCH http://localhost:8080/api/boards/507f1f77bcf86cd799439011 \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"name": "Updated Board Name"}'
```

### Delete Board
Delete a board.

- **Method**: `DELETE`
- **Path**: `/boards/:id`

#### Response (204 No Content)

#### Error Responses
- **400 Bad Request**: Invalid board ID or deletion error
  ```json
  {
    "error": "invalid id"
  }
  ```
- **401 Unauthorized**: Missing or invalid JWT token
  ```json
  {
    "error": "unauthorized"
  }
  ```
- **403 Forbidden**: User does not have access to the board

#### Example
```bash
curl -X DELETE http://localhost:8080/api/boards/507f1f77bcf86cd799439011 \
  -H "Authorization: Bearer <token>"
```

## Real-time Updates
Board creation, update, and deletion operations broadcast a `board_updated` event via Socket.IO to all connected clients in the "/" namespace. The event payload is `null`. This allows clients to refresh their board data in real-time.

## Notes
- Board access is restricted to owners and members.
- Members are specified as an array of user ID hex strings.
- Columns must have unique IDs within a board and include name and order.
- The `isArchived` field indicates if the board is archived (not modifiable in responses here).
- Timestamps (`createdAt`, `updatedAt`) are included in responses.