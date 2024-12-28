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

List all (or some) todo items. Supports simple text filtering. For example,
`godo list "+work"` will list all todos containing the substring "+work" in the
text.

Todos are displayed in an interactive numbered list. You can interact with the
list using the following commands:

- `q` or `quit`: Exit the interactive list
- `?` or `help`: Show help for the interactive list
- `d` or `delete`: Delete the selected todo item

More commands will be added in the future.

### `delete`

Delete a todo item by ID.

**Usage:**

```bash
godo delete [id]
```

**Arguments:**

- `id`: The ID of the todo to delete
