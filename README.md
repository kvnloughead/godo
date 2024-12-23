# GoDo

A todo tracking application written in Go. Features will eventually include:

1. A JSON Api to interact with, allowing CRUD operations on todo items.
2. User registration, activation, authentication, and authorization.
3. Todos stored in Postgresql.
4. A convenient command line interface for writing and viewing todos, utilizing a [todo.txt](https://github.com/todotxt/todo.txt) style syntax.
5. A CLI for interacting with the API.

- See [docs/api/SETUP.md](./docs/api/SETUP.md) for a quickstart guide.
- See [docs/cli/COMMANDS.md](./docs/cli/COMMANDS.md) for a list of commands.
- See [docs/api/ENDPOINTS.md](./docs/api/ENDPOINTS.md) for a list of API endpoints.

## Limitations

- Todo.txt style parsing is not implemented, so the context and projects fields are not filled. However searching by context and project is handled by text searches starting with `@` or `+`.

- Email activation is still in a development sandbox.
