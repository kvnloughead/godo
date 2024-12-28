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

List all (or some) todo items. Supports simple text filtering and enters an interactive mode for managing todos.

**Usage:**

```bash
godo list [pattern]
```

**Arguments:**

- `pattern`: Optional text to filter todos (e.g., "@phone" or "some task")

**Interactive Mode Commands:**

- `<number>`: Select a todo by its number
- `<number>rm|del|delete`: Delete the selected todo
- `<number>d|done`: Mark the selected todo as done
- `<number>u|undone`: Mark the selected todo as not done
- `<number>e|edit`: Edit the selected todo's text
- `<number>a|archive`: Archive the selected todo

**Other Commands:**

- `?` or `help`: Show help for interactive mode
- `q`, `quit`, or `exit`: Exit interactive mode

**Examples:**

````bash
# List all todos
godo list

# List todos containing "@phone"
godo list @phone

# In interactive mode:
1rm     # Delete todo #1
2d      # Mark todo #2 as done
3u      # Mark todo #3 as not done

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
````
