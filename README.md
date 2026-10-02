# cf-open

Open Cloudflare dashboard for your project from CLI.

## Installation

```bash
go install github.com/mst-mkt/cf-open/cmd/cf-open@latest
```

```bash
nix profile install github:mst-mkt/cf-open
```

## Usage

```bash
cf-open
```

This command reads your project's configuration (`cloudflare.config.ts`, `wrangler.jsonc`, `wrangler.json` or `wrangler.toml`) to list resources related to your project. The file is searched for in the current directory and then in its parent directories. You can select the resource you want to open, and its dashboard will open in your browser.

```bash
$ cf-open
? Select a resource to open:
  ▸ Worker: worker-name
    Observability: worker-name
    R2: bucket-name
    D1: database-name (database-id)
```

If there is only one resource, it will open directly.

Reading `cloudflare.config.ts` requires Node.js v22.18.0 or later, as the file is evaluated with `node` from your `PATH`. Use `--mode` to choose the mode passed to a function-form config.

The account ID in the dashboard URL is taken from the first of these that is set: `--account-id`, the configuration file, the `CLOUDFLARE_ACCOUNT_ID` environment variable, the account cached by `cf`, then the one cached by Wrangler.

### Options

| Option            | Description                                                                                                   |
| ----------------- | ------------------------------------------------------------------------------------------------------------- |
| `-c`, `--config`  | Path to the configuration file (`cloudflare.config.ts`, `wrangler.jsonc`, `wrangler.json` or `wrangler.toml`) |
| `-m`, `--mode`    | Mode passed to a function-form `cloudflare.config.ts`                                                         |
| `--account-id`    | Cloudflare account ID                                                                                         |
| `-a`, `--all`     | Open all resources in the browser                                                                             |
| `-p`, `--print`   | Print URL to stdout instead of opening in browser                                                             |
| `-v`, `--version` | Print the version number                                                                                      |

## Supported Resources

- Workers
- Pages
- Workers Observability
- Workers Cron Triggers
- Queues
- Workflows
- Browser Run
- VPC
- R2 Object Storage
- Worker KV
- D1 SQL Databases
- Pipelines
- Vectorize
- Secrets Store
- Images

## License

MIT License. See [LICENSE](LICENSE) for details.

## References

Inspired by and references to

- [`gh browse`](https://cli.github.com/manual/gh_browse)
- [`vercel open`](https://vercel.com/docs/cli/open)
- [m1guelpf/cf-url](https://github.com/m1guelpf/cf-url)
