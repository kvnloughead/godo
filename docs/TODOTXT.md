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
2. If the first word is a lowercase `x`, the task has been completed. A word that merely starts with `x`, like `xylophone`, is part of the body.
3. `(A)` represents a priority. It must appear immediately after the completion status. `A` through `Z` are allowed, with `A` having the highest priority. Lowercase letters, like `(a)`, are not priorities and stay in the body.
4. After the optional completion and priority fields, zero to two dates of YYYY-MM-DD format are allowed. There are three cases:
   - If a completed todo has any such dates, the first one is its completion date.
   - If a completed todo has two such dates, the second one is its creation date. The creation date must not be later than the completion date, or the line is rejected. The same day is allowed.
   - If an uncompleted todo has any such dates, the first one is its creation date.

   Any dates beyond these remain in the todo's text but are not parsed.

## The body

```
Text and +projects, @contexts, and key:value tags
```

1. Although not in the spec, we require that the body be non-empty.
2. Any text is allowed. If items matching any of the prefix formats are found in the body they are treated like plain text.
3. The body can optionally contain projects, contexts, and tags.

### Projects and contexts

1. A project is a word that starts with `+`, and a context is a word that starts with `@`. Because the whole word must match, `me@example.com` is plain text, not a context.
2. The `+` or `@` must be followed by one or more non-whitespace characters. All non-whitespace characters are acceptable. A lone `+` or `@` is plain text.
3. Projects and contexts stay in the todo's text. They are also collected into lists.

### Tags

1. Tags are key:value pairs.
2. The spec allows both key and value to contain _any_ non-whitespace characters aside from `:`. But this is very impractical. Consider how `https://google.com` would be parsed. Most likely, the author of the todo did not intend an `https` tag. So, we restrict tag keys and values to letters, digits, and underscores. Words that don't match are plain text.
3. Like projects and contexts, user tags stay in the todo's text and are also collected into a map. If a key appears more than once, the last value wins.

### GoDo's tags

The keys `id` and `mod` are reserved for internal use. They are expected to always be placed at the end of the line, to avoid conflict with user-created tags of the same names. If either appear within the description, or if they are duplicated, the line is rejected.

1. `id:` is the todo's identifier. New todos get a ULID, like `01JB8Z3K4MQ7Y2R5T8V0WXCDEF`. Todos created before sync existed have a UUID instead. IDs are treated as opaque strings and are never case-normalized. Any value of 1 to 64 letters, digits, or hyphens is accepted.
2. `mod:` is the time of the last change, as an RFC 3339 timestamp, like `2026-10-25T14:02:11Z`.
3. A malformed `id:` or `mod:` value is an error, and the whole line is rejected.
