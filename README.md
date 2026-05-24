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
# ask anything
sage "why is my disk full"
sage "what's using port 3000"
cat error.log | sage "what's wrong here"   # pipe data in

# follow-up questions
sage "is postgres running"
sage "why is it slow"                      # auto-detected as follow-up
sage ask "how do I fix that" --follow-up   # explicit

# fix a problem
sage fix "my disk is full"                 # diagnose + suggest + apply fix

# monitor over time
sage watch "is my CPU spiking"             # polls every 30s
sage watch "is postgres running" -i 10     # custom interval

# history
sage history                               # list past questions
sage history --last                        # show full last answer
sage history --search "postgres"           # search by keyword
sage history --clear                       # clear all history

# config
sage config                                # show current settings
sage config --provider ollama              # switch provider
sage config --provider claude --api-key sk-ant-...
sage config --provider openai --api-key sk-...
sage config --model gpt-4o                 # switch model
sage config --list-models                  # list local Ollama models
sage config --reset                        # reset to defaults

# other
sage update                                # update to latest release
sage --version                             # show version
sage test "question"                       # run 3x, check consistency
sage --grade "question"                    # grade answer accuracy
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
│   ├── fix.go        # diagnose + suggest + apply fix
│   ├── watch.go      # continuous monitoring
│   ├── update.go     # self-update from GitHub releases
│   ├── history.go    # browse past Q&A
│   ├── config.go     # settings management
│   ├── grade.go      # answer accuracy grading
│   ├── test.go       # consistency testing
│   ├── completion.go # shell completions
│   └── welcome.go    # default screen
└── internal/
    ├── ai/           # Ollama, Claude, OpenAI backends
    ├── config/       # persistent config (~/.sage/config.json)
    ├── context/      # follow-up session context (~/.sage/context.json)
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
