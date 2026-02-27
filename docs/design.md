# telegram-bot-cli（Go）设计草案

## 目标

构建一个 Go 版本 `telegram-bot-cli`，先聚焦高频需求：

1. 通过 polling 在本地接收用户消息
2. 提供基础 bot 查询与发送消息能力
3. token 可由用户自行配置（home 目录或参数指定）

## 技术选型

- 语言：Go
- CLI：标准库 `flag`（先做轻量实现）
- HTTP：`net/http`
- 输出：JSON/JSONL

## Token 配置策略

优先级（高 -> 低）：

1. `--token`
2. `TG_BOT_TOKEN`
3. `--config /path/to/config.json`（默认 `~/.tgbot-cli/config.json`）

配置文件支持：

- 纯文本 token
- JSON profile 结构（`active_profile` + `profiles`）

## 首批命令（已实现）

- `tgbot updates listen --interval 3s --timeout 20`
- `tgbot bot me`
- `tgbot message send --chat-id ... --text ...`

## 本地接收用户消息方案

只使用 polling：

1. 可选调用 `deleteWebhook` 清理 webhook 冲突
2. 周期调用 `getUpdates`
3. 按 `update_id + 1` 推进 offset，避免重复消费
4. `--interval` 控制轮询间隔周期
5. `--format` 支持 `pretty`（默认）和 `jsonl` 输出

## 下一步

1. 增加 `bot commands set/get`
2. 增加 `message send` 的 parse mode 支持（Markdown/HTML）
3. 增加结构化日志和错误码
