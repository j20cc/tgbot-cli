package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/example/tgbot-cli/internal/config"
	"github.com/example/tgbot-cli/internal/polling"
	"github.com/example/tgbot-cli/internal/telegram"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "updates":
		runUpdates(os.Args[2:])
	case "bot":
		runBot(os.Args[2:])
	case "message":
		runMessage(os.Args[2:])
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func runUpdates(args []string) {
	if len(args) == 0 {
		fatal("usage: tgbot updates <listen|list> [flags]")
	}

	switch args[0] {
	case "listen":
		runUpdatesListen(args[1:])
	case "list":
		runUpdatesList(args[1:])
	default:
		fatal("usage: tgbot updates <listen|list> [flags]")
	}
}

func runUpdatesListen(args []string) {
	fs := baseFlagSet("updates listen")
	interval := fs.Duration("interval", 2*time.Second, "polling interval between requests")
	timeout := fs.Int("timeout", 20, "getUpdates long-poll timeout in seconds")
	offset := fs.Int64("offset", 0, "initial update offset")
	once := fs.Bool("once", false, "run only one polling cycle")
	deleteWebhook := fs.Bool("delete-webhook", true, "delete webhook before polling")
	outputFormat := fs.String("format", "pretty", "updates output format: pretty|jsonl")
	tokenOpt := registerTokenFlags(fs)

	_ = fs.Parse(args)
	client := mustClient(tokenOpt)
	poller := polling.New(client, polling.Options{
		Interval:      *interval,
		TimeoutSecond: *timeout,
		InitialOffset: *offset,
		DeleteWebhook: *deleteWebhook,
		Once:          *once,
		OutputFormat:  *outputFormat,
	})

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := poller.Run(ctx, os.Stdout, os.Stderr); err != nil {
		if errors.Is(err, context.Canceled) {
			return
		}
		fatalf("polling failed: %v", err)
	}
}

func runUpdatesList(args []string) {
	fs := baseFlagSet("updates list")
	limit := fs.Int("limit", 10, "max number of latest updates to print")
	timeout := fs.Int("timeout", 0, "getUpdates timeout in seconds (default 0 for one-shot)")
	offset := fs.Int64("offset", 0, "initial update offset")
	deleteWebhook := fs.Bool("delete-webhook", true, "delete webhook before listing")
	outputFormat := fs.String("format", "pretty", "updates output format: pretty|jsonl")
	tokenOpt := registerTokenFlags(fs)

	_ = fs.Parse(args)
	if *limit <= 0 {
		fatal("--limit must be greater than 0")
	}

	ctx := context.Background()
	client := mustClient(tokenOpt)
	if *deleteWebhook {
		fmt.Fprintln(os.Stderr, "[info] deleting webhook before listing...")
		if err := client.DeleteWebhook(ctx); err != nil {
			fatalf("delete webhook failed: %v", err)
		}
	}

	recent := make([]telegram.Update, 0, *limit)
	currentOffset := *offset
	for {
		updates, err := client.GetUpdatesWithLimit(ctx, currentOffset, *timeout, 100)
		if err != nil {
			fatalf("updates list failed: %v", err)
		}
		if len(updates) == 0 {
			break
		}

		for _, update := range updates {
			recent = append(recent, update)
			if len(recent) > *limit {
				recent = recent[len(recent)-*limit:]
			}
			if update.UpdateID >= currentOffset {
				currentOffset = update.UpdateID + 1
			}
		}
	}

	for _, update := range recent {
		formatted, err := polling.FormatUpdate(update.Raw, *outputFormat)
		if err != nil {
			fatalf("format update failed: %v", err)
		}
		if _, err := os.Stdout.Write(formatted); err != nil {
			fatalf("write output failed: %v", err)
		}
	}
}

func runBot(args []string) {
	if len(args) == 0 {
		fatal("usage: tgbot bot <me> [flags]")
	}
	sub := args[0]
	switch sub {
	case "me":
		fs := baseFlagSet("bot me")
		tokenOpt := registerTokenFlags(fs)
		_ = fs.Parse(args[1:])
		client := mustClient(tokenOpt)
		res, err := client.GetMe(context.Background())
		if err != nil {
			fatalf("bot me failed: %v", err)
		}
		printJSON(res)
	default:
		fatal("usage: tgbot bot <me> [flags]")
	}
}

func runMessage(args []string) {
	if len(args) == 0 || args[0] != "send" {
		fatal("usage: tgbot message send --chat-id <id> --text <text> [flags]")
	}
	fs := baseFlagSet("message send")
	tokenOpt := registerTokenFlags(fs)
	chatID := fs.String("chat-id", "", "target chat id")
	text := fs.String("text", "", "message text")
	_ = fs.Parse(args[1:])
	if *chatID == "" || *text == "" {
		fatal("--chat-id and --text are required")
	}
	client := mustClient(tokenOpt)
	res, err := client.SendMessage(context.Background(), *chatID, *text)
	if err != nil {
		fatalf("message send failed: %v", err)
	}
	printJSON(res)
}

type tokenFlagOptions struct {
	token      *string
	configPath *string
	profile    *string
	apiBase    *string
}

func registerTokenFlags(fs *flag.FlagSet) tokenFlagOptions {
	return tokenFlagOptions{
		token:      fs.String("token", "", "telegram bot token"),
		configPath: fs.String("config", "", "config path (default ~/.tgbot-cli/config.json)"),
		profile:    fs.String("profile", "", "config profile name (defaults to active_profile)"),
		apiBase:    fs.String("api-base", "https://api.telegram.org", "telegram api base"),
	}
}

func mustClient(opts tokenFlagOptions) *telegram.Client {
	resolvedToken, err := config.ResolveToken(config.TokenOptions{
		TokenFlag:  *opts.token,
		ConfigPath: *opts.configPath,
		Profile:    *opts.profile,
	})
	if err != nil {
		fatalf("resolve token: %v", err)
	}
	return telegram.NewClient(*opts.apiBase, resolvedToken)
}

func baseFlagSet(name string) *flag.FlagSet {
	return flag.NewFlagSet(name, flag.ExitOnError)
}

func printJSON(raw []byte) {
	var out any
	if err := json.Unmarshal(raw, &out); err != nil {
		fmt.Println(string(raw))
		return
	}
	formatted, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		fmt.Println(string(raw))
		return
	}
	fmt.Println(string(formatted))
}

func printUsage() {
	fmt.Print(`tgbot - Telegram Bot CLI

Usage:
  tgbot updates listen [flags]
  tgbot updates list [flags]
  tgbot bot me [flags]
  tgbot message send --chat-id <id> --text <text> [flags]

Example:
  tgbot updates listen --interval 3s --timeout 20 --format pretty
  tgbot updates list --limit 20 --format pretty
  tgbot bot me
  tgbot message send --chat-id 12345 --text "hello"

Token resolution order:
  1) --token
  2) TG_BOT_TOKEN
  3) config file (~/.tgbot-cli/config.json)
`)
}

func fatal(msg string) {
	fmt.Fprintln(os.Stderr, msg)
	os.Exit(1)
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
