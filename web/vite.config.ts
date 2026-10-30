import vue from "@vitejs/plugin-vue";
import { defineConfig } from "vite";

// The dev server proxies API requests to the Go API, so the browser sees one
// origin in development just as it does in production, where the API serves
// the built files. Set GODO_API_URL to point at a different API.
const apiURL = process.env["GODO_API_URL"] ?? "http://localhost:4000";

// https://vite.dev/config/
export default defineConfig({
  plugins: [vue()],
  server: {
    proxy: {
      "/v1": apiURL,
    },
  },
});
