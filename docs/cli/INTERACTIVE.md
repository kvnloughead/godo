# Interactive Mode

The `godo list` command enters an interactive mode by default, allowing you to manage your todos using simple commands. Each todo is displayed with a number that can be used to reference it in commands.

## Display Format

Todos are displayed in two sections:

```
Todos:
 1. [ ] Active todo
 2. [✓] Completed todo

Archived:
 3. [ ] Archived todo
 4. [✓] Completed archived todo
```

## Commands

Commands follow the format: `command number [number...]`

### Basic Commands

- `done 1 2` - Mark todos #1 and #2 as completed
- `undone 3` - Mark todo #3 as not completed
- `rm 4 5` - Delete todos #4 and #5
- `archive 6` - Archive todo #6
- `unarchive 7` - Unarchive todo #7

### Command Aliases

- `delete`: `rm`, `del`
- `done`: `d`, `complete`
- `undone`: `u`, `undo`, `incomplete`
- `archive`: `a`
- `unarchive`: `ua`

### Other Commands

- `?` or `help` - Show help message
- `q` or `quit` - Exit interactive mode

## Examples

```bash
# Mark multiple todos as done
done 1 2 3

# Delete a single todo
rm 4

# Archive multiple todos
archive 5 6 7

# Mark todo as incomplete using alias
u 8
```
