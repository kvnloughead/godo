# The Todo.txt Format

This app's [todo parser](../internal/todo/parse.go) follows a fairly strict reading of the [todo.txt guidelines](https://github.com/todotxt/todo.txt). Here is a diagram of the format:

```
x (A) 2026-02-01 2026-01-01 Text and +projects, @contexts, and key:value tags
```

This can be divided into two pieces.

## The prefix

```
x (A) 2026-02-01 2026-01-01
```

1. Each space-separated "word" in the prefix is optional, with a partial exception involving the dates.
2. If an `x` is present as the first character, the task has been completed.
3. `(A)` represents a priority. It must appear immediately after the completion-status. `A` through `Z` are allowed, with `A` having the highest priority.
4. After the optional completion and priority fields, zero to two dates of YYYY-MM-DD format are allowed. There are three cases:
   - If a completed todo has any such dates, the first one is its completion date.
   - If a completed todo has two such dates, the second one is its creation date. However, the creation date must be earlier than the completion date.
   - If an uncompleted todo has any such dates, the first one is its creation date. Additional dates remain in the todo's text but are not parsed.

## The body

```
Text and +projects, @contexts, and key:value tags
```

1. Although not in the spec, we require that the body be non-empty.
2. Any text is allowed. If items matching any of the prefix formats are found in the body they are treated like plain text.
3. The body can optionally contain projects, contexts, and tags.

### Projects and contexts

1. A project starts with a `+` and a context starts with a `@`.
2. They can both be followed by one or more non-whitespace characters. All non-whitespace characters are acceptable.

### Tags

1. Tags are key:value pairs.
2. The spec allows both key and value to contain _any_ non-whitespace characters aside from `:`. But this is very impractical. Consider how `https:google.com` would be parsed. Most likely, the author of the todo did not intend an `http` tag. So, we are restricting tag keys and values to alphanumeric characters, numbers, and underscores. Tags that don't match this will be left in the todo text, but not parsed as tags.
3. The tags `mod` and `id` are reserved for internal use.
