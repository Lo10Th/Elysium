# Anthropic Emblem

Elysium emblem for the [Anthropic API](https://docs.anthropic.com) — covering Claude message creation and model listing for the Claude family of AI models.

---

## Prerequisites

You need an Anthropic API key. Export it as an environment variable:

```bash
export ANTHROPIC_API_KEY=sk-ant-...
```

> **Note:** Never commit your API key to source control. Use a `.env` file or a secrets manager in production.

The Anthropic API also requires an `anthropic-version` header on every request. The emblem defaults this to `2023-06-01`, which is the stable production version.

---

## Installation

```bash
# Install Elysium CLI
pip install elysium-cli   # or follow repo installation instructions

# Verify the emblem is registered
ely list
```

---

## Usage Examples

### 1. `create-message` — Create a Message with Claude

Send a conversation to Claude and receive a generated reply. This is the primary endpoint for all text generation tasks.

```bash
# Simple single-turn question
ely execute anthropic create-message \
  --model "claude-3-5-sonnet-20241022" \
  --messages '[{"role":"user","content":"What is the capital of France?"}]' \
  --max_tokens 1024

# With a system prompt for context and persona
ely execute anthropic create-message \
  --model "claude-3-5-sonnet-20241022" \
  --messages '[{"role":"user","content":"Review this Python function for bugs:\n\ndef add(a, b):\n    return a - b"}]' \
  --system "You are an expert Python engineer. Be concise and focus on correctness." \
  --max_tokens 512

# Multi-turn conversation
ely execute anthropic create-message \
  --model "claude-3-5-sonnet-20241022" \
  --messages '[
    {"role":"user","content":"My name is Alice."},
    {"role":"assistant","content":"Hello Alice! How can I help you today?"},
    {"role":"user","content":"What is my name?"}
  ]' \
  --max_tokens 256

# Lower temperature for deterministic output
ely execute anthropic create-message \
  --model "claude-3-haiku-20240307" \
  --messages '[{"role":"user","content":"Summarize the water cycle in 3 bullet points."}]' \
  --max_tokens 300 \
  --temperature 0.2

# Stop at a custom sequence
ely execute anthropic create-message \
  --model "claude-3-5-sonnet-20241022" \
  --messages '[{"role":"user","content":"List 10 fruits, one per line."}]' \
  --max_tokens 200 \
  --stop_sequences '["6."]'
```

**Key parameters:**

| Parameter | Type | Required | Description |
|---|---|---|---|
| `model` | string | ✅ | Claude model ID (e.g. `claude-3-5-sonnet-20241022`) |
| `messages` | array | ✅ | Conversation history (objects with `role` and `content`) |
| `max_tokens` | integer | ✅ | Maximum output tokens to generate |
| `system` | string | ❌ | System prompt for context and instructions |
| `temperature` | number | ❌ | Randomness (0–1). Default: `1` |
| `top_p` | number | ❌ | Nucleus sampling. Use temperature _or_ top_p, not both |
| `top_k` | integer | ❌ | Sample from top K tokens only |
| `stream` | boolean | ❌ | Stream response as server-sent events. Default: `false` |
| `stop_sequences` | array | ❌ | Sequences that trigger generation to stop |

**Recommended Claude models:**

| Model ID | Description |
|---|---|
| `claude-3-5-sonnet-20241022` | Best balance of intelligence and speed — recommended for most tasks |
| `claude-3-5-haiku-20241022` | Fastest and most compact Claude 3.5 model |
| `claude-3-opus-20240229` | Most capable for complex reasoning and nuanced tasks |
| `claude-3-haiku-20240307` | Ultra-fast, cost-efficient for high-throughput workloads |

---

### 2. `list-models` — List Available Claude Models

Retrieve all Claude models available to your API key.

```bash
ely execute anthropic list-models
```

**Example response:**
```json
{
  "data": [
    {
      "id": "claude-3-5-sonnet-20241022",
      "display_name": "Claude 3.5 Sonnet",
      "created_at": "2024-10-22T00:00:00Z"
    },
    {
      "id": "claude-3-5-haiku-20241022",
      "display_name": "Claude 3.5 Haiku",
      "created_at": "2024-10-22T00:00:00Z"
    },
    {
      "id": "claude-3-haiku-20240307",
      "display_name": "Claude 3 Haiku",
      "created_at": "2024-03-07T00:00:00Z"
    }
  ],
  "has_more": false,
  "first_id": "claude-3-5-sonnet-20241022",
  "last_id": "claude-3-haiku-20240307"
}
```

---

## Authentication Details

Anthropic uses a custom `x-api-key` header (not the standard `Authorization: Bearer` header). The emblem handles this automatically using the `ANTHROPIC_API_KEY` environment variable.

Every request must also include an `anthropic-version` header specifying the API version. The emblem defaults this to `2023-06-01`.

```
x-api-key: sk-ant-...
anthropic-version: 2023-06-01
Content-Type: application/json
```

---

## Error Handling

| Code | Error | Cause | Resolution |
|---|---|---|---|
| `400` | `bad_request` | Invalid request body, missing required fields, malformed messages | Review request structure — ensure `messages` alternates user/assistant roles |
| `401` | `unauthorized` | Missing or invalid API key | Verify `ANTHROPIC_API_KEY` is set and active |
| `429` | `rate_limit_exceeded` | Request or token rate limit exceeded | Implement exponential backoff; check your usage limits |
| `500` | `internal_error` | Server-side error on Anthropic's infrastructure | Retry after a short delay; check [status.anthropic.com](https://status.anthropic.com) |
| `529` | `overloaded` | Anthropic API temporarily overloaded due to high demand | Retry with backoff — this is transient and usually resolves in seconds |

**Example retry logic with backoff:**
```python
import time
import anthropic

client = anthropic.Anthropic()

for attempt in range(5):
    try:
        message = client.messages.create(
            model="claude-3-5-sonnet-20241022",
            max_tokens=1024,
            messages=[{"role": "user", "content": "Hello, Claude!"}]
        )
        print(message.content)
        break
    except anthropic.RateLimitError:
        wait = 2 ** attempt          # 1s, 2s, 4s, 8s, 16s
        print(f"Rate limited, retrying in {wait}s...")
        time.sleep(wait)
    except anthropic.APIStatusError as e:
        if e.status_code == 529:
            wait = 5 * (attempt + 1)
            print(f"API overloaded, retrying in {wait}s...")
            time.sleep(wait)
        else:
            raise
```

**Message validation tips:**
- Messages must alternate between `user` and `assistant` roles
- The first message must have role `user`
- `system` is a top-level parameter, not a message role (unlike OpenAI)
- `max_tokens` is **required** — there is no default

---

## Rate Limits

Anthropic enforces rate limits per API key across two dimensions:

| Dimension | Description |
|---|---|
| **RPM** (Requests per minute) | Maximum API calls within a 60-second window |
| **ITPM** (Input tokens per minute) | Maximum input tokens processed per minute |
| **OTPM** (Output tokens per minute) | Maximum output tokens generated per minute |

**Default Tier 1 limits (approximate):**

| Model | RPM | Input TPM | Output TPM |
|---|---|---|---|
| Claude 3.5 Sonnet | 50 | 40,000 | 8,000 |
| Claude 3.5 Haiku | 50 | 50,000 | 10,000 |
| Claude 3 Haiku | 50 | 50,000 | 10,000 |

> Limits increase as your account usage history grows. Check your current limits at [console.anthropic.com/settings/limits](https://console.anthropic.com/settings/limits).

**Best practices:**
- Prompt caching reduces costs and latency for repeated context (system prompts, document analysis)
- Use streaming (`stream: true`) for long responses to avoid timeout issues
- Set `max_tokens` conservatively — you're charged for output tokens generated
- Use `claude-3-haiku` for high-throughput, lower-stakes tasks to stay within rate limits

---

## Further Reading

- [Anthropic API Reference](https://docs.anthropic.com/en/api)
- [Messages Guide](https://docs.anthropic.com/en/docs/messages)
- [Claude Model Overview](https://docs.anthropic.com/en/docs/models-overview)
- [Prompt Engineering Guide](https://docs.anthropic.com/en/docs/prompt-engineering)
- [Rate Limits](https://docs.anthropic.com/en/docs/rate-limits)
- [Pricing](https://www.anthropic.com/pricing)
