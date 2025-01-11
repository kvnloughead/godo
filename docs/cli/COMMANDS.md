# CLI Commands

## User Management

### `register`

Register a new user account. Once registered, the user will receive an email with
an activation token. The user must activate their account before they can use the
CLI.

**Usage:**

```bash
godo register [-e email]
```

**Flags:**

- `-e, --email`: Email address for registration

**Example:**

```bash
godo register -e user@example.com
```

### `activate`

Activate a registered user account using the activation token sent via email.

**Usage:**

```bash
godo activate -t token
```

**Flags:**

- `-t, --token`: Activation token

### `auth`

Authenticate and create a session for subsequent commands. Email and password
can be provided via flags or prompted for securely.

**Usage:**

```bash
godo auth [-e email] [-p password]
```

**Flags:**

- `-e, --email`: Email address (optional, will prompt if not provided)
- `-p, --password`: Password (optional, will prompt securely if not provided)

## Todo Management

### `add`

Add a new todo item.

**Usage:**

```bash
godo add "Todo text here"
```

### `list`

List and manage todo items. By default, enters an interactive mode for managing todos.

**Usage:**

```bash
godo list [flags] [pattern]
```

**Arguments:**

- `pattern`: Optional text to filter todos (e.g., "@phone" or "some task")

**Flags:**

- `-p, --plain`: Output in plain text format (disables interactive mode)
- `--include-archived`: Include archived todos in the list
- `--only-archived`: Show only archived todos
- `-d, --done`: Show only completed todos
- `-u, --undone`: Show only incomplete todos

**Examples:**

```bash
# List all unarchived todos in interactive mode
godo list

# List todos containing "@phone" in interactive mode
godo list @phone

# List all todos (including archived) in plain text format
godo list --include-archived --plain
```

See [INTERACTIVE.md](./INTERACTIVE.md) for details about interactive mode.

### `delete`

Delete a todo item by ID.

**Usage:**

```bash
godo delete [id]
```

**Arguments:**

- `id`: The ID of the todo to delete

### `done`

Mark a todo item as completed.

**Usage:**

```bash
godo done [id]
```

### `undone`

Mark a todo item as not completed.

**Usage:**

```bash
godo undone [id]
```
