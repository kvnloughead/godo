import type { Backend } from "./backend";
import type { Filter, Todo } from "./types";

/** HttpBackend implements Backend against the /v1 API. The web app is served
from the same origin as the API (and the Vite dev server proxies /v1), so
requests use relative URLs and need no CORS. */
export class HttpBackend implements Backend {
  private readonly baseURL: string;

  constructor(baseURL = "/v1") {
    this.baseURL = baseURL;
  }

  list(_filter?: Filter): Promise<Todo[]> {
    return notImplemented(`${this.baseURL}/todos`);
  }

  get(uid: string): Promise<Todo> {
    return notImplemented(`${this.baseURL}/todos/${uid}`);
  }

  add(_text: string): Promise<Todo> {
    return notImplemented(`${this.baseURL}/todos`);
  }

  update(todo: Todo): Promise<Todo> {
    return notImplemented(`${this.baseURL}/todos/${todo.uid}`);
  }

  delete(uid: string): Promise<void> {
    return notImplemented(`${this.baseURL}/todos/${uid}`);
  }
}

function notImplemented(path: string): Promise<never> {
  return Promise.reject(new Error(`HttpBackend: ${path} not implemented`));
}
