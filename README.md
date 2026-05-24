<p align="center">
  <img src="sage-banner.png" alt="sage" width="100%" />
</p>

> Ask your system anything in plain English.

sage is an open source CLI tool that answers system questions using real data from your machine — not generic advice. It runs the right commands, reads the output, and explains it clearly.

```
sage "why is my disk full"
sage "what's using port 3000"
sage "is postgres running"
sage "how is my memory looking"
```

![sage demo](demo.gif)

---

## Why sage?

When something goes wrong on your machine, you either Google it and get generic answers, or you dig through commands you half-remember. sage does the digging for you — it figures out which commands to run, runs them, and explains what it found using the actual output from your system.

It's not a chatbot. It's a system assistant that reads your machine directly.

---

## How it works

```
you ask a question
  → AI decides which commands to run
  → sage runs them on your machine
  → AI explains the real output in plain English
```

sage never guesses. Every answer is backed by actual system data.

---

## Install

**macOS / Linux (one-liner):**

```sh
curl -fsSL https://raw.githubusercontent.com/yanonymousV2/sage/main/install.sh | sh
```

**From source:**

```sh
git clone https://github.com/yanonymousV2/sage
cd sage
go build -o sage .
sudo mv sage /usr/local/bin/sage
```

**Uninstall:**

```sh
curl -fsSL https://raw.githubusercontent.com/yanonymousV2/sage/main/uninstall.sh | sh
```

---

## Usage

```sh
sage "question"            # ask anything about your system
sage ask "question"        # same thing, explicit subcommand
sage history               # browse past questions
sage history --last        # show full last answer
sage history --search "pg" # search past answers
sage config                # show current settings
sage config --provider claude --api-key sk-ant-...
sage config --model gpt-4o
sage config --reset        # back to defaults
sage test "question"       # run 3x, check consistency
sage --grade "question"    # grade answer accuracy after response
```

---

## AI backends

sage works with local and cloud models.

| Provider | Default model | Needs API key |
|----------|--------------|---------------|
| Ollama (default) | qwen2.5:14b | No — runs locally |
| Claude | claude-sonnet-4-6 | Yes — console.anthropic.com |
| OpenAI | gpt-4o-mini | Yes — platform.openai.com |

**Switch provider:**

```sh
sage config --provider ollama          # local, free
sage config --provider claude --api-key sk-ant-...
sage config --provider openai --api-key sk-...
```

---

## Running locally with Ollama

```sh
brew install ollama
ollama pull qwen2.5:14b
ollama serve

sage "how is my system doing"
```

---

## Safety

sage only runs read-only commands. Before running anything, it shows you exactly what it plans to execute and asks for approval. Destructive commands are blocked entirely.

---

## Contributing

Pull requests are welcome.

**Run locally:**

```sh
git clone https://github.com/yanonymousV2/sage
cd sage
go mod tidy
go run main.go "how is my system doing"
```

**Project structure:**

```
sage/
├── main.go
├── cmd/
│   ├── ask.go        # core flow — command planning + explanation
│   ├── config.go     # sage config command
│   ├── history.go    # sage history command
│   ├── grade.go      # answer accuracy grading
│   ├── test.go       # consistency testing
│   └── welcome.go    # default screen
└── internal/
    ├── ai/           # Ollama, Claude, OpenAI backends
    ├── config/       # persistent config (~/.sage/config.json)
    ├── executor/     # shell command runner
    ├── history/      # history store (~/.sage/history.json)
    └── safety/       # blocked and dangerous command rules
```

**Adding a new backend:**

Implement the `Backend` interface in `internal/ai/`:

```go
type Backend interface {
    Complete(model, prompt string) (string, error)
    Stream(model, prompt string, onToken func(string)) (string, error)
}
```

Then register it in `internal/ai/backend.go`.

---

## Design

- **System-aware** — uses real data, never generic advice
- **Transparent** — shows every command before running it
- **Safe** — read-only by default, explicit approval for anything sensitive
- **Universal** — single binary, macOS and Linux, no runtime dependencies
- **Bring your own key** — open source, you control the AI backend

---

## License

MIT
