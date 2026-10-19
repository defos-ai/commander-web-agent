# Commander Web Agent

A lightweight [Commander](https://github.com/defos-ai/commander) agent for website
reviews using HTTP and static HTML. No browser or model API key is required.

| Action | Result |
| --- | --- |
| `discover-pages-no-browser` | Same-hostname pages and proposals for HTTP checks. |
| `discover-pages` | The same crawl, also proposing checks for the browser agent. |
| `check-links` | Broken links, HTTP statuses and skipped coverage. |
| `check-seo` | Metadata, headings, canonicals and indexing directives. |
| `detect-technologies` | Passive indicators from headers, metadata and script URLs. |

## Quick start

From this checkout, start Commander and this agent:

```sh
docker compose up --build --wait && docker compose run --rm --no-deps playground
```

Ask: **Run review-website-no-browser for https://defos.ai.**

Admin UI: <http://localhost:4285>. API: `http://localhost:4284`.
Get the API login with `docker compose exec api commander-api credentials`.
Stop the demo with `docker compose down`; add `--volumes` to delete its results.

The demo uses Commander **v0.1.0** and local filesystem providers. Set
`COMMANDER_VERSION` to choose a different Commander release.

## Connect to an existing Commander

Use the controller's bus and artifact storage:

```sh
BUS_URL=/shared/bus STORAGE_URL=/shared/storage go run ./cmd
```

For distributed deployments, use RabbitMQ (`BUS_URL=amqp://...` or `amqps://...`)
and S3-compatible storage (`STORAGE_URL=s3://bucket`). Supply provider credentials
through the environment. The agent needs no database.

`--agent-id-suffix=0123abcd` overrides `COMMANDER_AGENT_ID_SUFFIX`. The resulting
ID is `web-agent-0123abcd`; omit both for a fresh ID. Never reuse a suffix between
running instances of this type. See `go run ./cmd --help` for other options.

## Scope

Crawls stay on the exact hostname, allowing HTTP/HTTPS and port changes. They
inspect up to ten pages; link checks inspect up to 40 links. Public, internal and
localhost sites are supported. Metadata, link-local and reserved addresses are
blocked, DNS results are pinned per action, and environment proxies are ignored.

Static checks do not render JavaScript. Technology indicators are observations,
not verified runtime use or vulnerability findings. Pair with the
[browser agent](https://github.com/defos-ai/commander-browser-agent) for rendered
accessibility and screenshots; the controller schedules all proposed work.

## Development and releases

This repository follows the
[example-agent structure](https://github.com/defos-ai/commander-example-agent):
`agent.go`, `actions/`, `cli/`, `tests/`, `missions/` and `docker/`.

```sh
go test -race -count=1 -timeout=3m ./...
go vet ./...
```

The next agent release is **v0.2.0**, using Commander framework/providers **v0.1.0**.
The manual **Make release** workflow publishes native binaries, checksums,
`agent.yaml` and `ghcr.io/defos-ai/commander-web-agent:v0.2.0`. Existing image names
and previous tags are retained. Before its first release, grant this repository
write access to the organization's existing GHCR package. Complete releases
promote Docker `stable` and GitHub `Latest` when they are the newest version.

See [AGENTS.md](AGENTS.md) for integration checks and release details.
Agent source: [MIT](LICENSE). Commander framework and provider dependencies
remain AGPL-3.0-only. Their license requirements apply to combined builds.
