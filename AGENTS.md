# Working on the web agent

## Boundaries and structure

- Keep README short and user-facing; this guide owns implementation/test details.
  Keep one root `.gitignore` and a companion AGENTS.md for every README.
  Preserve uncommitted work, license notices and credentials. Do not commit or
  publish without authorization.
- Agent source is [MIT](LICENSE); Commander framework and provider dependencies
  remain AGPL-3.0-only. Preserve third-party license notices and the README
  distinction between source and dependency licenses for combined builds.
- This is a standalone module using published Commander framework/providers
  v0.1.0. Docker/CI use `GOWORK=off`; never add sibling `replace` directives or
  imports of Commander apps. Agent v0.2.0 does not require framework v0.2.0.
- [agent.go](agent.go) defines literal `AgentType = "web-agent"` and `NewAgent`.
  [actions.go](actions.go) creates fresh action instances; preserve explicitly
  empty catalogs. `cli/factory.go` chooses FS/RabbitMQ/S3 providers; the agent
  does not use a database or managed secrets encryption key.
- `cli.Run` owns the host lifetime without process signals. `cli.Main` translates
  process signals; close providers only after the runtime stops. Preserve joined
  provider errors during cancellation.
- `--agent-id-suffix` overrides `COMMANDER_AGENT_ID_SUFFIX`, including explicit
  empty. Accept exactly eight lowercase hexadecimal characters; empty generates
  a fresh ID. Validate the composed ID with v0.1.0 `id.RuntimeType` before opening
  providers. Never accept an external type prefix or expose the setting in
  public config. Identity correlates spawner retries; it is not authorization or
  duplicate prevention.
- Commander → Controller → Agent: actions propose work, controllers dispatch it.
  Agents upload bytes; controllers persist accepted artifact metadata. Do not
  execute follow-up proposals directly or claim exactly-once execution.

## HTTP scope and proposals

- `actions/http.go` bounds response parsing and exact-hostname scope. Schemes and
  ports may differ, but subdomains and `www` variants remain outside scope.
- `actions/network.go` validates every resolved IP and pins the hostname to those
  addresses for the action. Private/loopback targets are intentional;
  metadata/link-local/reserved addresses and environment proxies are blocked.
  Keep numeric-IP dialing and redirect hostname checks together.
- Preserve page/link limits and excluded coverage. Proposals use resolved URLs
  and the target action schemas. `discover-pages` proposes HTTP and browser
  actions; `discover-pages-no-browser` proposes only this agent's HTTP checks.
  Mission preflight and controller scheduling determine executor availability.
- SEO evidence is bounded. Technology versions are passive, unverified
  observations; redact credentials, queries and fragments from URL evidence.
- [The demo mission](missions/review-website-no-browser.json) runs without a
  browser. The client assigns discovery, reviews its proposed checks, then
  completes the operation. Creating an operation alone does not queue discovery.

## Compose and release

- Compose follows commander-example-agent: isolated FS state owned by UID1000,
  one controller, API/MCP/Admin bound to loopback ports4284/4283/4285. Commander
  images default to v0.1.0 (override `COMMANDER_VERSION`); their currently
  published platform is linux/amd64, requested explicitly. This agent image
  supports linux/amd64 and linux/arm64.
- MCP credentials live in a dedicated demo volume; only MCP and Playground read
  it. Retain configured overrides, never log tokens, and keep API/MCP at PID1 for
  credential lookup. Playground is optional and must not use `--all-in-one`.
- The demo fixes API `ENV` to `development`.
  `.github/composes/compose.test.yaml` enables test-only API cleanup and adds an
  unexposed HTTP fixture for the integration scenario. Keep fixtures and test-only
  services in that override, not the demo. Tests need no public target or model;
  Docker lifecycle stays in setup/CI, outside Go test fixtures.
- Release assets derive from the repository owner and literal AgentType in
  agent.go. Keep `.github/` and `agent.yaml.tmpl` aligned with the example agent.
  Agent packages describe delivery assets only: binaries/checksums and a Docker
  digest, never runtime secrets, arbitrary commands or host permissions.
- Publish from one tested revision. Keep build/push reusable-only, preserve
  immutable version tags/assets, serialize releases and advance stable only
  after complete publication with no backward version movement. Grant this new
  repository Actions write access to the existing GHCR package before release.
  Commander must stop publishing this agent's image during the ownership handoff.

## Validation

Use Go1.26.7 or newer. Put new tests in `tests/` as external packages; prefer
public API integration for cross-host behavior. `tests/network_test.go` exercises address rejection, DNS pinning, mixed answers
and proxy bypass through public actions with a local DNS fixture. Do not run
resolver-mutating tests in parallel or expose implementation helpers for tests.

```sh
GOWORK=off go test -race -count=1 -timeout=3m ./... ./.github/tests
GOWORK=off go vet ./... ./.github/scripts ./.github/tests
```

The integration test cleans all public operation data: use a disposable project
with ports4283–4285 free, never a development/shared API.

```sh
export COMPOSE_PROJECT_NAME=commander-web-test API_USERNAME=commander
export API_PASSWORD="$(openssl rand -hex 32)"
docker compose -f compose.yaml -f .github/composes/compose.test.yaml up --build --wait --wait-timeout 120
docker compose -f compose.yaml -f .github/composes/compose.test.yaml run --rm --no-deps -T playground --help
COMMANDER_API_URL=http://127.0.0.1:4284 GOWORK=off go test -tags=integration -race -count=1 -timeout=3m -run '^TestAgentThroughAPI$' -v ./tests/...
docker compose -f compose.yaml -f .github/composes/compose.test.yaml down --volumes --remove-orphans
unset API_PASSWORD API_USERNAME COMPOSE_PROJECT_NAME
```

Always tear down the test stack after failures too. The test waits for a ready
registration, queues discovery against its local fixture, checks automatic
follow-ups/results and completes a report, then settles and cleans work.

An ignored go.work may link this module and Commander framework/providers for
local development, but standalone validation must use published v0.1.0 modules.
