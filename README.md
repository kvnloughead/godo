# GoDo

A todo tracking application written in Go. Features will eventually include:

1. A JSON Api to interact with, allowing CRUD operations on todo items.
2. User registration, activation, authentication, and authorization.
3. Todos stored in Postgresql.
4. A convenient command line interface for writing and viewing todos, utilizing a [todo.txt](https://github.com/todotxt/todo.txt) style syntax.

See [SETUP.md](./SETUP.md) for a quickstart guide.

## Limitations

- Todo.txt style parsing is not implemented, so the context and projects fields are not filled. However searching by context and project is handled by text searches starting with `@` or `+`.

## Endpoints that don't require authentication

Usage example s assume you are running the app locally with either `make run/api` or `make run/air`. The API is currently not deployed.

### GET /v1/healthcheck

Displays application information, including the time and hash of the most recently made commit. If changes have been made since the last commit, the version has the string '-dirty' appended. Requires no permissions.

```bash
# Example usage
curl localhost:4000/v1/healthcheck
```

```json
// Example response
{
  "status": "available",
  "system_info": {
    "environment": "development",
    "version": "2024-05-26T23:49:46Z-c663c2e35824b8f2b6f776768ee22022d1e86163-dirty"
  }
}
```

### POST /v1/users

Registers a new user. The request's body must contain JSON with three fields: email, password, and name. Emails must be valid and unique. Password must be between 8 and 72 characters.

When the request is successful an email is sent to the user for activation instructions.

```bash
# Example usage
curl -X POST -d '{ "email": "user@mail.com", "password": "password", "name": "username" }' localhost:4000/v1/users
```

```json
// Example response
{
  "user": {
    "id": 1,
    "created_at": "2024-05-27T13:10:30-04:00",
    "name": "username",
    "email": "user@mail.com",
    "activated": false
  }
}
```

### PUT /v1/users/activation

Activates a user's account. The request's body must contain a token field with a valid token. The token is provided in an email sent upon registration, but it expires within three days. A new token can be issued with the `POST /v1/tokens/activation` endpoint.

```bash
# Example usage
curl -X PUT -d '{ "token": "PCXEWRH2WX6DSQIAPVBE24CY6I" }' localhost:4000/v1/users/activation
```

The response has a 202 status code and its body contains a message and the user object.

```json
// Example response
{
  "message": "user successfully activated",
  "user": {
    "id": 1,
    "created_at": "2024-05-27T13:10:30-04:00",
    "name": "username",
    "email": "user@mail.com",
    "activated": true
  }
}
```

### POST /v1/tokens/activation

Generates a new activation token and sends it in an email. The request's body
must contain the user's email.

```bash
# Example usage
curl -X POST -d '{ "email": "user@mail.com" }' localhost:4000/v1/tokens/activation
```

```json
// Example response
{
  "message": "an email will be sent to you containing activation instructions"
}
```

### GET /v1/tokens/authentication

Authenticates a user if the provided credentials are correct. The request body
must contain the user's email and password. The response contains an
authentication token. This authentication token should be used to authorize
all protected resources.

```bash
# Example usage
curl -X POST -d '{ "email": "user@mail.com", "password": "password" }' localhost:4000/v1/tokens/authentication
```

```json
// Example response
{
  "authentication_token": {
    "token": "EZVNRJHUXXXXXXZQGKTXIWDDFQ",
    "expiry": "2024-05-28T13:22:23.711932495-04:00"
  }
}
```

## Endpoints that require authentication

### GET /v1/todos

Returns an array containing the user's todo items, as well as some pagination data.
Supports text search via the `text` query parameter.

```bash
# List all todos
curl -H "Authorization: Bearer EZVNRJHUXXXXXXZQGKTXIWDDFQ" localhost:4000/v1/todos

# Search todos (matches anywhere in the text)
curl -H "Authorization: Bearer EZVNRJHUXXXXXXZQGKTXIWDDFQ" localhost:4000/v1/todos?text=milk

# Search by context
curl -H "Authorization: Bearer EZVNRJHUXXXXXXZQGKTXIWDDFQ" localhost:4000/v1/todos?text=@home

# Search by project
curl -H "Authorization: Bearer EZVNRJHUXXXXXXZQGKTXIWDDFQ" localhost:4000/v1/todos?text=%2Bshopping
```

```json
// Example response
{
  "paginationData": {
    "current_page": 1,
    "page_size": 20,
    "first_page": 1,
    "last_page": 1,
    "total_records": 2
  },
  "todos": [
    {
      "id": 1,
      "text": "buy milk @home +shopping"
      // ... other fields ...
    }
    // ... more todos ...
  ]
}
```

Note: When using curl directly, the `+` symbol in projects needs to be URL encoded as `%2B`.
The CLI handles this encoding automatically.

### POST /v1/todos

Add a new todo to the table. The request body must contain a text field with stores the text of the todo item. This is the only required field.

```bash
# Example usage
curl -X POST -H "Authorization: Bearer F6SB76ZCLKLJBHP7K7A6N2S7JM" -d '{ "text": "(A) do something important @readme +godo" }' localhost:4000/v1/todos

```

```json
// Example response
{
  "todo": {
    "id": 1,
    "user_id": 1,
    "created_at": "2024-05-27T17:44:39-04:00",
    "text": "(A) do something important @readme +godo",
    "priority": 0,
    "completed": false,
    "version": 1
  }
}
```

### GET /v1/todos/:id

Retrieves a todo by its ID, but only if it is owned by the current user.

If there is no such todo a 404 response is sent.

```bash
# Example usage
curl -X POST -H "Authorization: Bearer F6SB76ZCLKLJBHP7K7A6N2S7JM" localhost:4000/v1/todos/1
```

```json
// Example response
{
  "todo": {
    "id": 1
    // ...
  }
}
```

### PATCH /v1/todos/:id

Updates the todo with the provided ID, but only if it is owned by the current user.

If there is no such todo a 404 response is sent.

```bash
# Example usage
curl -X PATCH -H "Authorization: Bearer F6SB76ZCLKLJBHP7K7A6N2S7JM" -d '{ "completed": true }' localhost:4000/v1/todos/1
```

```json
// Example response
{
  "todo": {
    "id": 1,
    // ...
    "completed": true,
    "version": 2 // version will be incremented
  }
}
```

### DELETE /v1/todos/:id

Deletes a todo by its ID, but only if it is owned by the current user.

If there is no such todo a 404 response is sent.

```bash
# Example usage
curl -X DELETE -H "Authorization: Bearer F6SB76ZCLKLJBHP7K7A6N2S7JM" localhost:4000/v1/todos/1
```
