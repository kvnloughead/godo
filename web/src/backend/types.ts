/** Types shared by every Backend adapter. They mirror the JSON returned by
/v1/todos today (internal/server/data.Todo). The Wails adapter will need
todo.Todo to marshal to the same shape (see docs/SPEC.md, Web app). */

export interface Todo {
  uid: string;
  text: string;
  priority: string;
  completed: boolean;
  archived: boolean;
  projects?: string[];
  contexts?: string[];
  created_at: string;
  updated_at: string;
  version: number;
}

/** Mirrors store.Filter. Omitted fields use the Go zero values: unarchived
todos, any completion status. */
export interface Filter {
  text?: string;
  archived?: "exclude" | "include" | "only";
  completion?: "any" | "done" | "undone";
}

/** Mirrors store.ErrNotFound, ErrConflict, and ErrUnavailable, so components
can handle failures the same way whichever adapter is in use. */
export class NotFoundError extends Error {}
export class ConflictError extends Error {}
export class UnavailableError extends Error {}
