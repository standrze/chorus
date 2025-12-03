## Chorus

Chorus is a tiny **multi‑agent playground** written in Go.
It wires up a couple of AI “agents” on top of an OpenAI‑compatible API (for example a local `llama.cpp` server) and is meant as a sandbox for **agents talking to each other** from a simple CLI.

Right now the default command brings up two agents:

- **Teacher agent** – writes a short beginner Polish lesson / dialogue.
- **Professor agent** – reviews and comments on the teacher’s output.

The logic for a real back‑and‑forth conversation is intentionally minimal so you can customize it however you like.

---

### Quick start

- **Clone and build**:

```bash
git clone https://github.com/standrze/chorus.git
cd chorus
go build -o chorus .
```

- **Run your model server**  
Start an OpenAI‑compatible endpoint (the sample code assumes something like a `llama.cpp` server) at:

```text
http://localhost:12434/engines/llama.cpp/v1
```

- **Run Chorus**:

```bash
./chorus
```

This will create a client, construct the two agents and a `Conversation`, and call `Conversation.Start` with a prompt for a short Polish dialogue.

> **Note:** `Conversation.Start` is currently just a stub that validates there are at least two agents. It’s the main place you’ll extend the project to actually orchestrate multi‑turn agent chatter.

---

### How it’s put together

- **`main.go`** – Minimal entrypoint that calls `cmd.Execute()`.
- **`cmd/root.go`** – Cobra root command (`chorus`); sets up the OpenAI client, two example agents, and kicks off the conversation.
- **`pkg/agent/agent.go`** – Core `Agent` type:
  - wraps the OpenAI Go client,
  - stores messages, tools, and options,
  - exposes a `Send` method for chat completions,
  - includes helpers like `WithModel`, `WithReasoningEffort`, `WithFunctionTools`, etc.
- **`pkg/agent/name.go`** – `GenerateAgentName()` utility (Docker‑style `adjective_noun` name generator).
- **`pkg/agent/conversation.go`** – `Conversation` struct holding a context, a slice of agents, and a `maxTurns` limit; `Start` is the hook for your custom orchestration.
- **`internal/app.go`** – Placeholder `App` and `Config` types for future expansion.

---

### Configuration hints

Right now almost everything is hard‑coded in `cmd/root.go`:

- **Base URL**:
  - `option.WithBaseURL("http://localhost:12434/engines/llama.cpp/v1")`
- **Model**:
  - `"ai/gpt-oss"`
- **Reasoning effort**:
  - `openai.ReasoningEffortMedium`

To adapt this to your own setup:

- Point the base URL at your own server or OpenAI endpoint.
- Swap in a different model name supported by that endpoint.
- Change the system prompts for each agent to give them different personalities / roles.

---

### Ideas for extending

- Implement a proper **turn‑taking loop** in `Conversation.Start`, feeding one agent’s output into the next.
- Add more roles (planner, executor, critic, summarizer, etc.) and wire them into the conversation.
- Promote the hard‑coded options in `cmd/root.go` into **flags or a config file**.
- Add more Cobra subcommands for different agent setups.

---

### License

See the `LICENSE` file in the repo root for full terms.


