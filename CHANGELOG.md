# Changelog

## Unreleased

- Added `tgbot updates list` for one-shot retrieval of latest N updates via `--limit`.
- Added `GetUpdatesWithLimit` in Telegram client to support bounded `getUpdates` requests.

## v0.1.0

First runnable version for trial:

- `tgbot updates listen` with polling-only local update consumption and configurable `--interval`.
- `tgbot bot me` to inspect bot identity.
- `tgbot message send` to send plain text messages.
- Token resolution: `--token` > `TG_BOT_TOKEN` > `~/.tgbot-cli/config.json`.
