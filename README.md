# tgbot-cli

A lightweight Go CLI for Telegram bot local debugging with polling.

## Version

Current trial version: `v0.1.0`.

## Build

```bash
go build .
```

## Commands

- `tgbot updates listen` - poll updates and print JSON (pretty by default)
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

## Get bot identity

```bash
./tgbot-cli bot me --token <token>
```

## Send a message

```bash
./tgbot-cli message send --chat-id <chat-id> --text "hello" --token <token>
```
