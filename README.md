# cf-open

Open Cloudflare dashboard for your project from CLI.

## Installation

```bash
go install github.com/mst-mkt/cf-open@latest
```

## Usage

```bash
cf-open
```

This command reads Wrangler configuration (e.g., `wrangler.jsonc` or `wrangler.toml`) to list resources related to your project. You can select the resource you want to open, and its dashboard will open in your browser.

```bash
cf-open
? Select a resource to open:
  ▸ Worker: worker-name
    Observability: worker-name
    R2: bucket-name
    D1: database-name (database-id)
```

If there is only one resource, it will open directly.

### Options

| Option              | Description                                                               |
| ------------------- | ------------------------------------------------------------------------- |
| `--wrangler-config` | Path to the wrangler configuration file. Supports JSONC and TOML formats. |
| `--account-id`      | Cloudflare account ID                                                     |
| `-a`, `--all`       | Open all resources in the browser                                         |
| `-v`, `--version`   | Print the version number                                                  |

## Supported Resources

- Workers
- Workers Observability
- Workers Cron Triggers
- Queues
- Workflows
- Browser Rendering
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
