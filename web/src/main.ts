import { createApp } from "vue";
import "./style.css";
import App from "./App.vue";
import { backendKey, createBackend } from "./backend";

createApp(App).provide(backendKey, createBackend()).mount("#app");
