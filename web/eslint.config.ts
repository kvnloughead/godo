import skipFormatting from "@vue/eslint-config-prettier/skip-formatting";
import {
  defineConfigWithVueTs,
  vueTsConfigs,
} from "@vue/eslint-config-typescript";
import { globalIgnores } from "eslint/config";
import pluginVue from "eslint-plugin-vue";

// The strictest presets available: typescript-eslint's strict and stylistic
// type-checked rules (they use type information, so they catch things like
// floating promises and unsafe any that tsc doesn't), and eslint-plugin-vue's
// full recommended set. Prettier owns formatting, so skipFormatting turns off
// rules that would conflict with it.
export default defineConfigWithVueTs(
  {
    name: "app/files-to-lint",
    files: ["**/*.{ts,mts,tsx,vue}"],
  },
  globalIgnores(["dist/**", "coverage/**"]),
  pluginVue.configs["flat/recommended"],
  vueTsConfigs.strictTypeChecked,
  vueTsConfigs.stylisticTypeChecked,
  {
    name: "app/rules",
    rules: {
      // Match tsc's noUnusedParameters: a leading underscore marks a name as
      // intentionally unused.
      "@typescript-eslint/no-unused-vars": [
        "error",
        {
          argsIgnorePattern: "^_",
          varsIgnorePattern: "^_",
          caughtErrorsIgnorePattern: "^_",
        },
      ],
    },
  },
  skipFormatting,
);
