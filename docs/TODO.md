# TODO

## Fixes

- [ ] Fix panic when deleting multiple todos

## CLI

### Commands

- [ ] `edit` command to modify existing todos
- [ ] Command to sync edited todos back to API
- [x] `done` command to mark todos as completed
- [x] Command to mark todos as not done ()
- [x] Interactive listing with numbered results

### Error Handling

- [ ] CLI should return authentication error responses appropriately

### Todo Listing and Filtering

- [ ] Add archived status to output of --plain
- [ ] Display creation and modification dates
- [ ] Display IDs
- [ ] Support for due dates
- [ ] Priority system (todo.txt style)
- [ ] DRY commands in list.go
- [x] variadic arguments for commands (not just in interactive mode`)
- [x] document list command more fully
- [x] Archive management (archive/unarchive/view archive)
- [x] Interactive selection of todos by number

## Interactive Mode

- [ ] add pagination
- [ ] add search
- [ ] add sorting
- [ ] add filtering
- [ ] add command history
- [ ] add command completion
- [ ] log commands immediately, not after interactive mode is exited
- [ ] improve visibility of "unknown command" types of interactive mode errors
- [ ] re-prompt for todos if none provided
- [ ] re-prompt for command if none provided
- [ ] generate command help from command struct tags
- [ ] Toggle to show/hide completed tasks
- [ ] Check/uncheck all
- [ ] Prompt before deleting todos

---

## API

### Security Enhancements and Logging

- [ ] Implement SSH authentication
- [ ] Log suspicious patterns separately
- [ ] Track repeated failed attempts
- [x] Add rate limiting information

### Token Management

- [ ] Handle activation tokens in token package

### Deployment

- [ ] Leave the sandbox email environment when deploying to prod
