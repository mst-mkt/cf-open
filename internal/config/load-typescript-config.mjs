import { writeSync } from "node:fs";
import * as nodeModule from "node:module";
import { fileURLToPath, pathToFileURL } from "node:url";

const DEFINITION = Symbol.for("@cloudflare/config:definition");
const WORKER_TYPE = "cf-worker";
const WORKER_SCHEME = "cf-worker:";

const RESULT_FD = 3;

const isRecord = (value) => typeof value === "object" && value !== null;
const unwrap = async (value, ctx) => (typeof value === "function" ? value(ctx) : value);
const reasonOf = (error) => (error instanceof Error ? error.message : String(error));

// Type stripping is unflagged from v22.18.0.
const isSupportedRuntime = () => {
  if (process.versions.bun !== undefined) return false;
  if (typeof nodeModule.registerHooks !== "function") return false;

  const [major, minor] = process.versions.node.split(".").map((n) => Number.parseInt(n, 10));

  return major > 22 || (major === 22 && minor >= 18);
};

// Resolves `with { type: 'cf-worker' }` imports to the entrypoint path without loading it.
const registerConfigHooks = () =>
  nodeModule.registerHooks({
    resolve: (specifier, context, nextResolve) => {
      if ((context.importAttributes ?? {}).type !== WORKER_TYPE) return nextResolve(specifier, context);

      const isRelative = specifier.startsWith("./") || specifier.startsWith("../");
      const entrypoint =
        isRelative && context.parentURL
          ? fileURLToPath(new URL(specifier, context.parentURL))
          : specifier;

      return {
        url: `${WORKER_SCHEME}${encodeURIComponent(entrypoint)}`,
        format: "module",
        shortCircuit: true,
      };
    },
    load: (url, context, nextLoad) => {
      if (!url.startsWith(WORKER_SCHEME)) return nextLoad(url, context);

      const entrypoint = decodeURIComponent(url.slice(WORKER_SCHEME.length));

      return {
        format: "module",
        source: `export default ${JSON.stringify(entrypoint)}`,
        shortCircuit: true,
      };
    },
  });

const resolveExport = async (value, ctx) => {
  if (!isRecord(value) || !(DEFINITION in value)) return unwrap(value, ctx);

  const { config, type } = value[DEFINITION];
  const resolved = await unwrap(config, ctx);

  return isRecord(resolved) ? { ...resolved, type } : resolved;
};

const resolveConfig = async (configPath, ctx) => {
  registerConfigHooks();

  const config = await import(pathToFileURL(configPath).href);
  const worker = await resolveExport(config.default, ctx);

  if (!isRecord(worker) || worker.type !== "worker") {
    throw new Error("the default export must be a worker");
  }

  const settings =
    config.settings === undefined ? undefined : await resolveExport(config.settings, ctx);

  return { worker, settings };
};

const main = async () => {
  const [configPath, mode] = process.argv.slice(2);

  if (configPath === undefined) {
    throw new Error("usage: load-typescript-config <config-path> [mode]");
  }

  if (!isSupportedRuntime()) {
    throw new Error("reading cloudflare.config.ts requires Node.js v22.18.0 or later");
  }

  const result = await resolveConfig(configPath, { mode }).catch((error) => {
    throw new Error(`failed to load ${configPath}: ${reasonOf(error)}`);
  });

  // fd 3 rather than stdout, which the config itself may write to.
  writeSync(RESULT_FD, JSON.stringify(result));

  // The config may hold the event loop open, so the result would never reach the caller.
  process.exit(0);
};

main().catch((error) => {
  // writeSync rather than process.stderr, which is asynchronous on macOS.
  writeSync(2, `${reasonOf(error)}\n`);
  process.exit(1);
});
