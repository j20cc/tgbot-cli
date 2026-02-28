# tgbot-cli

A lightweight Go CLI for Telegram bot local debugging with polling.

## Version

Current trial version: `v0.1.0`.

## Build

```bash
go build .
```

## Commands

- `tgbot updates listen` - continuous polling and streaming output
- `tgbot updates list` - one-shot fetch and print latest N updates
- `tgbot bot me` - show current bot profile
- `tgbot message send` - send a text message

## Token configuration

Token is resolved in this order:

1. `--token`
2. `TG_BOT_TOKEN`
3. `--config /path/to/file` (or default `~/.tgbot-cli/config.json`)

Config file supports either:

- plain text token, e.g. `123456:ABC...`
- JSON with profiles:

```json
{
  "active_profile": "dev",
  "profiles": {
    "dev": {"token": "123456:ABC..."}
  }
}
```

## Listen for updates with polling

```bash
./tgbot-cli updates listen --interval 3s --timeout 20 --format pretty
```

Useful flags:

- `--interval`: interval between polling rounds
- `--timeout`: Telegram `getUpdates` timeout (seconds)
- `--offset`: initial update offset
- `--once`: run a single polling round and exit
- `--delete-webhook`: delete webhook before polling (default `true`)
- `--format`: output format, `pretty` (default) or `jsonl`

## List latest updates once

```bash
./tgbot-cli updates list --limit 20 --format pretty
```

Useful flags:

- `--limit`: number of latest updates to keep and print
- `--offset`: starting offset if you want to continue from a known update id
- `--delete-webhook`: delete webhook before listing (default `true`)
- `--timeout`: getUpdates timeout in seconds (default `0` for snapshot)
- `--format`: output format, `pretty` (default) or `jsonl`

## Get bot identity

```bash
./tgbot-cli bot me --token <token>
```

## Send a message

```bash
./tgbot-cli message send --chat-id <chat-id> --text "hello" --token <token>
```
