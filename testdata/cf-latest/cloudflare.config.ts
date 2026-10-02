import { bindings, defineConfig, defineWorker, exports, triggers } from "cf/config";
import * as entrypoint from "./src/index.ts" with { type: "cf-worker" };

const worker = defineWorker((ctx) => ({
  name: `cf-open-check-${ctx.mode ?? "default"}`,
  entrypoint,
  compatibilityDate: "2026-09-30",
  observability: { enabled: true },
  triggers: [
    triggers.fetch({ pattern: "example.com/*", zone: "example.com" }),
    triggers.scheduled({ schedule: "0 * * * *" }),
  ],
  exports: {
    MyWorkflow: exports.workflow({ name: "my-workflow" }),
  },
  env: {
    BROWSER: bindings.browser(),
    BUCKET: bindings.r2({ name: "my-bucket" }),
    CACHE: bindings.kv({ id: "kv-id" }),
    DB: bindings.d1({ name: "my-db", id: "db-id" }),
    EU_BUCKET: bindings.r2({ name: "my-eu-bucket", jurisdiction: "eu" }),
    IMAGES: bindings.images(),
    INDEX: bindings.vectorize({ name: "my-index" }),
    PIPELINE: bindings.pipeline({ name: "my-pipeline" }),
    QUEUE: bindings.queue({ name: "my-queue" }),
    SECRET: bindings.secretsStoreSecret({ storeId: "store-id", secretName: "my-secret" }),
    VPC: bindings.vpcService({ id: "vpc-id" }),
    WORKFLOW: bindings.workflow({ name: "remote-workflow", worker: "other-worker", exportName: "RemoteWorkflow" }),
  },
}));

export default defineConfig({
  accountId: "0123456789abcdef0123456789abcdef",
  worker,
});
