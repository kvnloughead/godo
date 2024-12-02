# GoDo CLI Development Guide

## Installation

To install the CLI globally:

```bash
make cli/install
```

This will build and install the `godo` command to `/usr/local/bin`.

## Development

### Building Locally

To build without installing:

```bash
make cli/build
```

The binary will be created in the current directory. You can run it directly with:

```bash
./godo [command]
```

### Configuration

The CLI can be configured in several ways (in order of precedence):

1. Command-line flags:

   ```bash
   godo --config /path/to/custom/config.json
   ```

2. Environment variables:

   ```bash
   export GODO_API_URL="http://localhost:4000/v1"
   ```

3. Configuration file at `~/.config/godo/settings.json`:

   ```json
   {
     "api_base_url": "http://localhost:4000/v1"
   }
   ```

4. Default values

### Making Changes

1. Make your changes to the code
2. Rebuild and install:
   ```bash
   make cli/install
   ```
3. Test your changes:
   ```bash
   godo [command]
   ```

### Viewing Logs

To view the CLI's log files in your default editor:

```bash
make cli/logs
```

### Available Settings

| Setting      | Description               | Environment Variable | Default                  |
| ------------ | ------------------------- | -------------------- | ------------------------ |
| api_base_url | Base URL for the GoDo API | GODO_API_URL         | http://localhost:4000/v1 |

## Project Structure

```
cmd/cli/
├── cmd/            # Command implementations
│   ├── root.go     # Base command and app initialization
│   ├── add.go      # Add todo command
│   └── ...         # Other commands
├── config/         # Configuration management
│   └── config.go
└── main.go         # Entry point
```

## Adding New Commands

1. Create a new command file in `cmd/cli/cmd/`:

   ```bash
   cobra add [command-name]
   ```

2. Implement the command logic in the new file
3. The command will automatically be added to the root command
