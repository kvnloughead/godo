# GoDo web app

The GoDo frontend: Vue 3, TypeScript, and Vite. The same build is meant to run in the browser and, later, in the Wails desktop app. The design is in the [Web app section of the spec](../docs/SPEC.md#web-app).

## Layout

```
web/
├── src/
│   ├── main.ts          # creates the app and provides the Backend
│   ├── App.vue
│   └── backend/
│       ├── backend.ts   # Backend interface (mirrors store.Store)
│       ├── types.ts     # Todo, Filter, and error types
│       ├── http.ts      # HttpBackend: calls /v1 (web)
│       └── index.ts     # backendKey and createBackend
├── embed.go             # package web: embeds dist/ with -tags embedweb
└── noembed.go           # without the tag: no frontend embedded
```

Components get the backend with `inject(backendKey)` and never call `fetch` or Wails directly.

## Scripts

Run these from `web/`, or use the `make web/*` targets from the repo root.

| Command                | What it does                                       |
| ---------------------- | -------------------------------------------------- |
| `npm run dev`          | Dev server on :5173, proxying `/v1` to :4000       |
| `npm run build`        | Type check with `vue-tsc`, then build into `dist/` |
| `npm run typecheck`    | Type check only                                    |
| `npm test`             | Run Vitest once                                    |
| `npm run format`       | Format with Prettier                               |
| `npm run format:check` | Check formatting (run by CI)                       |

## How the API serves it

`go build -tags embedweb ./cmd/api` embeds `dist/` in the API binary. Run `npm run build` first. Paths that match no API route are served from `dist/`, and unknown paths get `index.html` so client-side routes work. The Docker image builds the frontend and embeds it. `make web/api` does the same thing locally.
