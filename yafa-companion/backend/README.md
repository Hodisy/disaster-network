# Yafa Go backend — personal agent harness

This directory contains the separate Go backend requested for Yafa's VPS. It does not replace Disaster Network's Nitro backend or connect the existing mobile prototype to a server.

## Existing project conventions

Disaster Network already uses thin endpoints under `server/api`, use cases under `server/core`, database connectors under `server/repositories`, and external connectors under `server/services`. The Yafa frontend remains JavaScript / Nuxt: pages bind composables to props/events presentation components. Its present data is local to `useMobileCompanion.js`.

The Go counterpart follows the same dependency direction:

```text
HTTP / Live streaming connector
  -> core use case
     -> repository interface -> actual database connector
     -> service interface    -> external / audio connector

entities <- shared by these layers

Nuxt page -> composable / API connector -> Go HTTP interface
```

The firmware architecture and proposed trust model are documented in `../firmware/architecture.md`. The future transport layer must derive owner/device identity from validated credentials before invoking cores; caller-supplied owner IDs are not authentication. Each device has its own credential, mobile clients use their own user sessions, and the VPS private TLS/provider keys stay on the server. Manual single-device pairing is now implemented for the Waveshare; mobile user sessions and multi-device enrollment remain unimplemented. The infrastructure API retains its separate private bootstrap bearer credential.

## Hexagonal contract — 25 October 2026

Inbound connectors (HTTP handlers, future cron callbacks, device events, webhooks or agent tool handlers) decode/validate transport input and invoke a core use case. “Trigger” names the event source; “handler” names its inbound adapter. Do not put business orchestration in either.

Cores contain internal use-case logic and depend on ports/interfaces, not HTTP clients, Strata or Zenoh. They call repositories for this system's stored data and services for external APIs/capabilities. Repositories expose domain methods and return domain entities; their concrete connectors map storage records into those entities. Entities can own pure invariant/behaviour methods, but do not fetch data or call external APIs themselves. Keep workflow orchestration in cores. The current generic `MemoryRecord` is an infrastructure bridge only; it is not the final task/project repository design.

```text
HTTP / schedule / device / webhook / agent tool
  -> inbound handler
     -> core use case
        -> repository port -> storage connector -> entity
        -> service port -> external API connector
```

Realtime is an output/synchronization adapter: after committed changes, selected events reach the UI hub, which uses core/repository reads as needed and sends small updates. It is not another place for task/project business logic. The API executable is the composition root that wires real connectors into core ports.

This boundary makes future isolated unit tests possible; no unit tests are added or run unless the user explicitly requests them. No framework, dependency injection container or automatic multi-agent split is required.

## Implemented source

- Device identity, actual hardware capabilities, desired presentation state, audio format and voice-plan entities.
- A device repository interface and an audio transport interface. There is no mock repository or fake audio connector.
- `ReadDevice`: reads through the repository and checks record identity, owner and disabled state.
- `PrepareVoice`: requires an explicit start, actual microphone/speaker capabilities, a recent heartbeat, a supported audio format and an available transport.
- A validated voice plan is not an active session and never claims that recording or playback started.

The 25 October local infrastructure slice adds `cmd/api`, thin authenticated handlers, memory cores, a repository port and a real HTTP connector to the Strata owner service. It supports bounded infrastructure-record GET/POST/PATCH, version conflicts and health endpoints. These generic record routes remain distinct from the project/task/thread API. QUERY, DELETE and UI events deliberately return 501 rather than invented results. See [local stack instructions](../infra/README.md).

The [first agent harness](../docs/agent-harness-implementation.md) adds project/task/thread APIs, a bounded Responses loop and a server-side Live WebSocket connector. Native capture/playback, UI subscriptions, scheduling and frontend API wiring remain separate work. No paid model or voice call has been verified. The Go executable was built and started locally on 25 October 2026. Authenticated create/read/PATCH, stale-version rejection, and exact record/version recovery after all three containers were recreated succeeded against Strata. That initial infrastructure check wrote one harmless record; the later harness verification used isolated keys and did not seed the personal workspace. No unit tests, paid API call or deployment was performed. The original device/voice domain contracts below remain separate from the initial memory HTTP slice.

## Hardware boundary

The connected device enumerates as an Adafruit Feather ESP32-S2 TFT. The user's present unit has a display and Wi-Fi, with no microphone, speaker or attached battery. Its stored capabilities must therefore have Display=true, Microphone=false and Speaker=false. It must never advertise voice readiness. The ordered audio-capable device needs its exact board and audio codec identified before its driver/pins are selected.

The user selected C++ firmware, using the Arduino-ESP32 framework and Adafruit's existing display driver for the current board. Keep the board driver replaceable for the ordered audio device. CircuitPython was an earlier option and is no longer the selected implementation. TinyGo's currently documented native Wi-Fi support lists ESP32, C3 and S3 rather than this S2; no TinyGo compatibility is claimed here. See `../firmware/README.md` for the C++ firmware and real-time audio specification.

## Next implementation slices

1. Inspect the existing board firmware and preserve anything needed before replacing it. Display Hello Yafa and a small local animation.
2. Configure private Wi-Fi credentials locally, then retrieve desired presentation state over authenticated HTTPS. A desired state remains separate from the device's acknowledged/applied version.
3. Add real persistence for tasks, conversations, schedules, paired devices and command acknowledgements. Use record versions and explicit conflict checks; add persisted due jobs before calling anything a scheduled reminder.
4. Connect the implemented project/task/thread APIs to a frontend API connector. Keep frontend mock selection inside the composable, never inside presentation components or the Go backend.
5. Connect native phone capture/playback to the implemented Live transport, after an explicit user action. It can exercise the server voice path while the current home display receives expression/state changes.
6. For the audio-capable board, implement actual capture/playback buffers and a negotiated transport. Keep control messages separate from binary audio frames; acknowledge interruption, drain playback on cancel, reconnect cleanly and preserve local mute state.

For future audio, the protocol must distinguish unsupported, unavailable, preparing, listening, thinking, speaking, interrupted, disconnected and failed. Do not show listening or speaking based solely on an AI text response. A session cannot become active until the real transport acknowledges it. Provider credentials remain on the VPS; a paired device receives only its own scoped device credential.

Start with mono signed 16-bit PCM as an internal contract, negotiating the sample rate with the actual provider and hardware. This is not a claim about a particular OpenAI wire format. Echo cancellation, simultaneous capture/playback and hardware mute are not supplied by these domain interfaces.

## Sources

- [Adafruit Feather setup](https://learn.adafruit.com/adafruit-esp32-s2-tft-feather/circuitpython)
- [CircuitPython explanation](https://docs.circuitpython.org/en/latest/README.html)
- [TinyGo wireless support](https://github.com/tinygo-org/espradio)


## Physical Waveshare connection — 25 October 2026

[Device connection/setup](../firmware/DEVICE-CONNECTION.md) documents the authenticated local connection and future configurable HTTPS origin. The Waveshare now has verified local display/touch/microphone/speaker hardware; these capabilities do not establish an available remote voice stream. The older Feather still has no audio path.

The new device-only listener exposes `POST /v1/device/sync`, with a distinct configured device ID and 64-hex credential. Device reports are stored through `core.DeviceLink` → `DeviceLinkRepository` → `DeviceLinkMemory` → the existing real memory service. There is no backend fixture or invented fallback. Actual reports establish last-seen, boot ID and applied state; desired expression/display commands use persisted string versions and compare-and-write.

The separate admin listener adds `GET /v1/devices/{id}` and `PUT /v1/devices/{id}/presentation` (expected version, expression, explicit awake). Requested and applied versions remain separate. The listener available to the device cannot access admin records, start voice, control Wi-Fi or execute arbitrary tools. The optional `cmd/device-relay` bridges the Mac's LAN interface to Colima loopback during local development; a VPS HTTPS proxy replaces it later.

Go and firmware builds, real memory readiness and negative authentication/route checks passed. End-to-end physical results are recorded in the firmware verification report after configuration. No unit tests, agent loop, audio streaming, AI call or VPS deployment are part of this slice.

Observed connected slice: firmware v0.5.2 loaded saved Wi-Fi/server configuration after reboot, registered the physical Waveshare in real Strata, and acknowledged Happy version 2 and Idle version 3. A stale command returned 409. The corrected Setup portal opened/closed without resetting the device; a subsequent minute retained the same boot ID, connected Wi-Fi/Go, advancing exchanges and zero display errors. This is a bounded local check, not a public TLS, audio or long-term reliability test.

The user also physically removed/reconnected USB power. Go observed a different boot ID, restored online state and acknowledged version 3 without any new provisioning. The public VPS/HTTPS path remains untested.


## Independent audio bench

`cmd/audio-lab` is a separate native Mac executable for the [local USB audio lab](../firmware/AUDIO-LAB.md). It builds and runs without loading the main API configuration, memory service or bus. `scripts/start-audio-lab.py` creates the ignored `.env.audio-lab`; an OpenAI key is needed only for the explicit file-transcription step. Source MP3/WAV preparation and bounded level adjustment use local FFmpeg/Go, with no provider call. USB microphone/speaker streaming is handled by the local serial helper and firmware, with complete audio files kept only on the Mac.


## Voice-only local Live bench — 25 October 2026

Run `bun run audio` from `yafa-companion`, open `http://127.0.0.1:4180/audio`, and keep Yafa connected by USB. This starts only the isolated Go relay and USB page; no main application, database, bus or VPS is needed. A tap starts the voice connection, a second tap stops it; the device shows only the mascot. Live captions/session events are in the web page. `Ctrl+C` stops both services. The private key stays in `backend/.env.audio-lab`. See the firmware AUDIO-LAB guide for actual verification, the continuous microphone path, remaining acoustic echo limitations, the three-minute limit and actual verification.

## Shared Markdown prompt context

`core/prompt.go` loads the five allowlisted context files through the workspace repository, substitutes bounded `{{variables}}`, and snapshots instructions for both text runs and Live sessions. Seed templates are embedded from `prompts/`; startup fills missing documents in Strata before listening. Runtime reads and tool edits use only actual versioned Strata documents. Existing documents are never overwritten by initialization. See [the full prompt](../docs/agent-system-prompt.md). No new build, test or paid provider call verified this extension.

The startup initializer commits missing context files atomically using the current workspace version, retries only version conflicts, preserves unrelated state and refuses startup on storage/validation errors. It performs no write when all five files already exist. This change has not been built or run; actual initialization awaits the next updated API start.

## Agenda domain

`core/agenda.go` and `core/agenda_tools.go` implement the shared Me/Yafa timeline, Paris day/week boundaries, actual daily thread pointers and bounded conversation reads. `connectors/httpapi/agenda.go` is the thin read adapter. No business logic lives in presentation or HTTP handlers. See [the agenda contracts](../docs/agenda-and-tools.md); this source extension was not built or run.

## Typed commands and subjective daily state

`core/commands.go` exposes one strict `command({command: string})` function and generates compact Markdown help from internal domain schemas. `core/command_cli.go` parses up to four commands joined by unquoted &&; `core/command_batch.go` dispatches sequentially with explicit partial results, short errors and the existing loop limits. See [the CLI contract](../docs/command-cli.md). `core/daily_state.go` implements quoted, versioned subjective observations through the real Strata repository. Business rules remain in cores. See [the contract and unimplemented capability backlog](../docs/daily-state-and-commands.md). Source formatted only; no build or runtime check was performed.


## Hosted images and analysis

`core/hosted.go` dispatches the `images.generate` / `analysis.run` domains through a `services.HostedTools` port. The OpenAI connector uses Images generation and Responses Code Interpreter; execution never occurs on the VPS. Small generated files use `repositories.FileRepository` and immutable Strata chunks with authenticated metadata/content routes. The main loop retains one CLI function; nested hosted usage is tracked separately and analysis tokens also contribute to the parent budget. The private OPENAI_IMAGE_MODEL example/Compose mapping enables image setup; no actual environment or services were changed. See [configuration, bounds and remaining work](../docs/hosted-image-and-analysis.md). Source formatted/reviewed only; not built or run.


## Rolling quotas, costs and terminal verification

`core/safety.go` persists owner policy and rolling admission reservations through the real memory repository. `/v1/safety` supplies 5h/24h limits, usage/reservations and recent errors; confirmed PATCH edits use an expected revision. The model has no policy-edit tool. Main paid calls require YAFA_AI_ENABLED=true. `/v1/agent/commands/validate` is validation-only. Existing device awake=false uses the separate device presentation API.

`cmd/chat` is the thin interactive HTTP client. Supply YAFA_API_TOKEN privately through your shell/credential store, then run `YAFA_API_URL=http://127.0.0.1:8088 go run ./cmd/chat` from this backend folder. It needs no OpenAI key. `/usage` and `/status` are read-only; ordinary messages can incur costs only when the backend is configured, enabled and admits them. It reuses the same thread and requests cancellation on interruption without replaying a failed POST.

Native Go builds, the Docker API image and isolated real-store checks passed. No unit-test framework or paid model call was run. See [the verification report](../docs/backend-verification-2026-10-25.md), [cost tracking](../docs/backend-safety-and-costs.md) and [future iOS Settings contract](../docs/ios-settings-usage.md). Real provider/voice timing and mobile wiring remain unverified.


## Explicit optional capabilities and Codex jobs

Optional main-runtime capabilities require private environment opt-in: `YAFA_CODEX_ENABLED`, `YAFA_IMAGES_ENABLED`, `YAFA_ANALYSIS_ENABLED`, `YAFA_LIVE_ENABLED` and reserved `YAFA_SPOTIFY_ENABLED`, all false by default. Credentials alone no longer expose image/analysis commands. One filtered registry drives model help, validation and dispatch. Spotify remains absent until its OAuth connector exists.

`core/codex.go` owns asynchronous jobs/admission/state through `repositories.CodexRepository`. `connectors/services/codex_ssh.go` implements the service port; `cmd/codex-runner` normalizes remote CLI results. No callback or administrative Yafa secret is given to Codex. Remote setup is still required. See [configuration, accounting and verification](../docs/optional-tools-and-codex.md).

## Private personal-context migration

The separate `cmd/context-import` maintenance executable validates the private package, previews an additive merge and requires an explicit target version/fingerprint to commit it. Tasks, source metadata, bounded profile/preferences, on-demand documents and the idempotent receipt commit in one workspace compare-and-write. `memory.list` / `memory.read` expose metadata and bounded historical excerpts; only the short dated now.md orientation is injected by default. No personal bundle is embedded or public, and no import executes automatically at startup. See [usage and rollout requirements](../docs/personal-context-import.md).
