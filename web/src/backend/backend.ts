import type { Filter, Todo } from "./types";

/** Backend is the frontend's view of store.Store (internal/store). Components
call it and never talk to HTTP or Wails directly, so the same components
run in the browser (http.ts) and in the desktop app (wails.ts, later).*/
export interface Backend {
  list(filter?: Filter): Promise<Todo[]>;
  get(uid: string): Promise<Todo>;
  add(text: string): Promise<Todo>;
  update(todo: Todo): Promise<Todo>;
  delete(uid: string): Promise<void>;
}
