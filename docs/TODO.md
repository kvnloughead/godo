# TODO

### Todo Management Commands

- [ ] `done` command to mark todos as completed
- [ ] Command to mark todos as not done (name TBD)
- [ ] `update` command to modify existing todos
- [ ] Interactive listing with numbered results and pagination
- [ ] Command to edit todos in preferred editor
- [ ] Command to sync edited todos back to API

### Refactoring

- [] Handle activation tokens in token package

### Todo Listing and Filtering

- [ ] Toggle to show/hide completed tasks
- [ ] Archive management (archive/unarchive/view archive)
- [ ] Display creation and modification dates
- [ ] Support for due dates
- [ ] Priority system (todo.txt style)
- [ ] Interactive selection of todos by number

## GET /v1/todos

- [ ] Implement search by context (@context)
- [ ] Implement search by project (+project)
- [ ] Verify that all sort, filter, pagination options work
- [ ] In readme, describe options for sorting, filtering, and pagination
- [ ] Implement additional query params for filtering

## Todo Model

- [ ] Implement and validate todo metadata
- [ ] Add fields for:
  - [ ] completion status
  - [ ] creation date
  - [ ] modification date
  - [ ] due date
  - [ ] priority
  - [ ] archived status
  - [ ] contexts (array)
  - [ ] projects (array)

## Security Enhancements and Logging

- [ ] Implement SSH authentication
- [ ] Log suspicious patterns separately
- [ ] Track repeated failed attempts
- [x] Add rate limiting information

## Error Handling

- [ ] CLI should return authentication error responses appropriately

## Configuration

- [x] DRY up the config code for apiBaseURL and config files

## Development

- [x] Separate tokens for prod and dev
- [x] Increase token life for dev

## Deployment

- [ ] Leave the sandbox email environment when deploying to prod
