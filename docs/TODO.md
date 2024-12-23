# TODO

## CLI

- [x] Simplify the process for creating a new user (CLI command(s))
- [x] Prompt for email and password securely in the CLI
- [x] Hide password in the CLI

## GET /v1/todos

- [ ] Implement search by context
- [ ] Implement search by project
- [ ] verify that all sort, filter, pagination options work
- [ ] in readme, describe options for sorting, filtering, and pagination of GET request
- [ ] implement additional query params for filtering.
- [x] Implement search by text

## Documentation

- [ ] Specify the status codes for all responses

## Todo model

- [x] associate todos with a given user. Users will not see other users' todos.
- [ ] implement and validate todo metadata for todo items

## Security Enhancements and Logging

- [ ] Log suspicious patterns separately
- [ ] Track repeated failed attempts
- [x] Add rate limiting information

## Error handling

- [] cli should return authentication error responses appropriately. Try running an unauthorized command to see the problem.

## Configuration

- [x] DRY up the config code for apiBaseURL and config files

## Development

- [x] separate tokens for prod and dev
- [x] increase token life for dev

## Deployment

- [ ] Leave the sandbox email environment when deploying to prod
