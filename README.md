<div align="center">
  <img src="logo_rounded.png" alt="Chorus Logo" width="200" />

  <h1>Chorus</h1>

  <p>
    <b>Multi-Agent Playground for AI Conversations</b>
  </p>

  <p>
    <a href="https://go.dev/"><img src="https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat-square&logo=go" alt="Go Version" /></a>
    <a href="https://github.com/standrze/chorus/blob/main/LICENSE"><img src="https://img.shields.io/badge/License-MIT-green?style=flat-square" alt="License" /></a>
  </p>
</div>

<br />

> **Chorus** is a tiny, versatile **multi‑agent playground** written in Go. It wires up AI "agents" on top of an OpenAI‑compatible API (for example a local `llama.cpp` server) and serves as a sandbox for **agents talking to each other** from a simple CLI.

---

## ✨ Features

- 🎭 **Multi-Agent Conversations**: Orchestrate conversations between multiple AI agents with customizable roles and personalities
- 🔌 **OpenAI-Compatible API**: Works with any OpenAI-compatible endpoint, including local `llama.cpp` servers
- 🛠️ **Flexible Agent System**: Easy-to-use agent API with support for system messages, tools, and custom configurations
- 🎯 **Minimal & Extensible**: Intentionally minimal core logic that you can customize to build complex agent interactions
- 💻 **CLI Interface**: Simple command-line interface built with Cobra for easy usage and extension

### Current Default Agents

Right now the default command brings up two agents:

- 👨‍🏫 **Teacher agent** – writes educational content
- 👨‍🔬 **Professor agent** – reviews and comments on the teacher's output

The logic for a real back‑and‑forth conversation is intentionally minimal so you can customize it however you like.

## 🚀 Installation

### Prerequisites

- **Go** 1.21 or higher
- An **OpenAI-compatible API endpoint** (e.g., `llama.cpp` server)

### Install from Source

```bash
git clone https://github.com/standrze/chorus.git
cd chorus
go build -o chorus .
```

Or run directly:

```bash
go run .
```

---

## 📖 Usage

### Starting Chorus

Before running Chorus, make sure you have an OpenAI-compatible endpoint running. The sample code assumes a `llama.cpp` server at:

```
http://localhost:12434/engines/llama.cpp/v1
```

Then simply run:

```bash
./chorus
```

This will create a client, construct the two agents and a `Conversation`, and call `Conversation.Start` with a prompt to begin the conversation.

> 💡 **Note:** `Conversation.Start` is currently just a stub that validates there are at least two agents. It's the main place you'll extend the project to actually orchestrate multi‑turn agent chatter.

---

## 🏗️ Architecture

### Project Structure

| File | Description |
|------|-------------|
| **`main.go`** | Minimal entrypoint that calls `cmd.Execute()` |
| **`cmd/root.go`** | Cobra root command (`chorus`); sets up the OpenAI client, two example agents, and kicks off the conversation |
| **`pkg/agent/agent.go`** | Core `Agent` type that wraps the OpenAI Go client, stores messages, tools, and options, exposes a `Send` method for chat completions, and includes helpers like `WithModel`, `WithReasoningEffort`, `WithFunctionTools`, etc. |
| **`pkg/agent/name.go`** | `GenerateAgentName()` utility (Docker‑style `adjective_noun` name generator) |
| **`pkg/agent/conversation.go`** | `Conversation` struct holding a context, a slice of agents, and a `maxTurns` limit; `Start` is the hook for your custom orchestration |
| **`internal/app.go`** | Placeholder `App` and `Config` types for future expansion |

---

## ⚙️ Configuration

Right now almost everything is hard‑coded in `cmd/root.go`:

| Setting | Current Value |
|---------|--------------|
| **Base URL** | `option.WithBaseURL("http://localhost:12434/engines/llama.cpp/v1")` |
| **Model** | `"ai/gpt-oss"` |
| **Reasoning effort** | `openai.ReasoningEffortMedium` |

### Customization

To adapt this to your own setup:

- Point the base URL at your own server or OpenAI endpoint
- Swap in a different model name supported by that endpoint
- Change the system prompts for each agent to give them different personalities / roles

---

## 🔮 Extending Chorus

### Ideas for Future Development

- 🔄 Implement a proper **turn‑taking loop** in `Conversation.Start`, feeding one agent's output into the next
- 👥 Add more roles (planner, executor, critic, summarizer, etc.) and wire them into the conversation
- 📝 Promote the hard‑coded options in `cmd/root.go` into **flags or a config file**
- 🎛️ Add more Cobra subcommands for different agent setups

---

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](https://github.com/standrze/chorus/blob/main/LICENSE) file for details.


