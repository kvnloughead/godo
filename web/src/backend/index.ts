import type { InjectionKey } from "vue";
import type { Backend } from "./backend";
import { HttpBackend } from "./http";

export type { Backend } from "./backend";
export * from "./types";

/** backendKey is the injection key for the app's Backend. main.ts provides it,
and components get it with inject(backendKey).*/
export const backendKey: InjectionKey<Backend> = Symbol("backend");

/** createBackend picks the adapter for the current environment. Only the HTTP
adapter exists so far; the Wails desktop build will add one.*/
export function createBackend(): Backend {
  return new HttpBackend();
}
