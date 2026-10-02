import { writeFileSync, writeSync } from "node:fs";
import * as nodeModule from "node:module";
import { fileURLToPath, pathToFileURL } from "node:url";

const WORKER_TYPE = "cf-worker";
const WORKER_SCHEME = "cf-worker:";

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
      if ((context.importAttributes ?? {}).type !== WORKER_TYPE)
        return nextResolve(specifier, context);

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

const resolveConfig = async (configPath, ctx) => {
  registerConfigHooks();

  const config = await import(pathToFileURL(configPath).href);
  if (!("default" in config)) throw new Error("the config has no default export");

  const root = await unwrap(config.default, ctx);
  if (!isRecord(root)) throw new Error("the default export must be an object");

  if (root.accountId !== undefined && typeof root.accountId !== "string") {
    throw new Error("accountId must be a string");
  }

  const worker = root.worker === undefined ? undefined : await unwrap(root.worker, ctx);
  if (worker != null && !isRecord(worker)) throw new Error("the worker must be an object");

  const containers = Array.isArray(root.containers)
    ? (await Promise.all(root.containers.map((container) => unwrap(container, ctx)))).filter(isRecord)
    : undefined;

  return { accountId: root.accountId, worker, containers };
};

const main = async () => {
  const [resultPath, configPath, mode] = process.argv.slice(2);

  if (resultPath === undefined || configPath === undefined) {
    throw new Error("usage: load-typescript-config <result-path> <config-path> [mode]");
  }

  if (!isSupportedRuntime()) {
    throw new Error("reading cloudflare.config.ts requires Node.js v22.18.0 or later");
  }

  const result = await resolveConfig(configPath, { mode, isPreview: false }).catch((error) => {
    throw new Error(`failed to load ${configPath}: ${reasonOf(error)}`);
  });

  // A file rather than stdout, which the config itself may write to.
  writeFileSync(resultPath, JSON.stringify(result));

  // The config may hold the event loop open, so the result would never reach the caller.
  process.exit(0);
};

main().catch((error) => {
  // writeSync rather than process.stderr, which is asynchronous on macOS.
  writeSync(2, `${reasonOf(error)}\n`);
  process.exit(1);
});
