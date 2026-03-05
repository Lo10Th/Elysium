# OpenAI Emblem

Elysium emblem for the [OpenAI API](https://platform.openai.com/docs) — covering chat completions (GPT-4, GPT-3.5-turbo), model listing, text embeddings, and image generation (DALL-E 2 / DALL-E 3).

---

## Prerequisites

You need an OpenAI API key. Export it as an environment variable:

```bash
export OPENAI_API_KEY=sk-...
```

> **Note:** Never commit your API key to source control. Use a `.env` file or a secrets manager in production.

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

### 1. `chat-completions` — Create a Chat Completion

Send a conversation to GPT-4 or GPT-3.5-turbo and receive a generated reply.

```bash
# Basic GPT-4o chat request
ely execute openai chat-completions \
  --model "gpt-4o" \
  --messages '[{"role":"user","content":"Explain quantum entanglement in one paragraph."}]'

# With a system prompt and temperature tuning
ely execute openai chat-completions \
  --model "gpt-3.5-turbo" \
  --messages '[{"role":"system","content":"You are a concise technical writer."},{"role":"user","content":"What is a hash table?"}]' \
  --temperature 0.3 \
  --max_tokens 256

# Generate multiple completion alternatives
ely execute openai chat-completions \
  --model "gpt-4o" \
  --messages '[{"role":"user","content":"Give me a creative product name for a coffee brand."}]' \
  --n 3 \
  --temperature 1.2
```

**Key parameters:**

| Parameter | Type | Required | Description |
|---|---|---|---|
| `model` | string | ✅ | Model ID (e.g. `gpt-4o`, `gpt-3.5-turbo`) |
| `messages` | array | ✅ | Conversation history with `role` and `content` |
| `temperature` | number | ❌ | Randomness (0–2). Default: `1` |
| `max_tokens` | integer | ❌ | Max tokens to generate |
| `stream` | boolean | ❌ | Stream partial deltas. Default: `false` |
| `top_p` | number | ❌ | Nucleus sampling (0–1). Default: `1` |
| `frequency_penalty` | number | ❌ | Repetition penalty (-2 to 2). Default: `0` |
| `presence_penalty` | number | ❌ | Topic diversity penalty (-2 to 2). Default: `0` |
| `n` | integer | ❌ | Number of completions to generate. Default: `1` |
| `stop` | string | ❌ | Stop sequence(s) |

---

### 2. `list-models` — List Available Models

Retrieve all models your API key has access to.

```bash
ely execute openai list-models
```

**Example response (excerpt):**
```json
{
  "object": "list",
  "data": [
    { "id": "gpt-4o", "object": "model", "created": 1715367049, "owned_by": "system" },
    { "id": "gpt-3.5-turbo", "object": "model", "created": 1677610602, "owned_by": "openai" },
    { "id": "text-embedding-3-small", "object": "model", "created": 1705948997, "owned_by": "system" }
  ]
}
```

---

### 3. `create-embedding` — Generate Text Embeddings

Convert text into a numerical vector representation for semantic search, clustering, or similarity comparisons.

```bash
# Single string embedding
ely execute openai create-embedding \
  --model "text-embedding-3-small" \
  --input "The quick brown fox jumps over the lazy dog"

# High-dimensional embedding for more precise similarity
ely execute openai create-embedding \
  --model "text-embedding-3-large" \
  --input "OpenAI provides state-of-the-art language models via REST API"
```

**Key parameters:**

| Parameter | Type | Required | Description |
|---|---|---|---|
| `model` | string | ✅ | Embedding model (e.g. `text-embedding-3-small`, `text-embedding-ada-002`) |
| `input` | string | ✅ | Text to embed (string, array of strings, or token arrays) |

**Recommended models:**
- `text-embedding-3-small` — fast, efficient, cost-effective
- `text-embedding-3-large` — highest quality, larger vectors (3072 dimensions)
- `text-embedding-ada-002` — legacy, still widely used

---

### 4. `create-image` — Generate Images with DALL-E

Generate images from a text prompt using DALL-E 2 or DALL-E 3.

```bash
# Simple DALL-E 3 image (1024×1024)
ely execute openai create-image \
  --prompt "A serene Japanese zen garden at dawn, photorealistic" \
  --model "dall-e-3"

# High-definition wide landscape
ely execute openai create-image \
  --prompt "A futuristic city skyline at sunset with flying vehicles" \
  --model "dall-e-3" \
  --size "1792x1024" \
  --quality "hd" \
  --style "vivid"

# Multiple DALL-E 2 variations
ely execute openai create-image \
  --prompt "Abstract geometric art with vibrant colors" \
  --model "dall-e-2" \
  --n 3 \
  --size "512x512"

# Get base64 image data instead of a URL
ely execute openai create-image \
  --prompt "A cute robot reading a book" \
  --model "dall-e-3" \
  --response_format "b64_json"
```

**Key parameters:**

| Parameter | Type | Required | Description |
|---|---|---|---|
| `prompt` | string | ✅ | Text description of the image |
| `model` | string | ❌ | `dall-e-2` or `dall-e-3`. Default: `dall-e-2` |
| `n` | integer | ❌ | Number of images (1–10 for DALL-E 2, 1 for DALL-E 3). Default: `1` |
| `size` | string | ❌ | Image dimensions (see table below) |
| `quality` | string | ❌ | `standard` or `hd` (DALL-E 3 only). Default: `standard` |
| `response_format` | string | ❌ | `url` or `b64_json`. Default: `url` |
| `style` | string | ❌ | `vivid` or `natural` (DALL-E 3 only). Default: `vivid` |

**Supported sizes:**

| Size | DALL-E 2 | DALL-E 3 |
|---|---|---|
| `256x256` | ✅ | ❌ |
| `512x512` | ✅ | ❌ |
| `1024x1024` | ✅ | ✅ |
| `1792x1024` | ❌ | ✅ |
| `1024x1792` | ❌ | ✅ |

---

## Error Handling

| Code | Error | Cause | Resolution |
|---|---|---|---|
| `400` | `bad_request` | Invalid parameters, unrecognized model, prompt flagged by safety filters | Review request body and prompt content |
| `401` | `unauthorized` | Missing or invalid API key | Verify `OPENAI_API_KEY` is set and valid |
| `429` | `rate_limit_exceeded` | Requests per minute or token quota exceeded | Implement exponential backoff; check your usage tier |
| `500` | `internal_error` | Server-side error on OpenAI's infrastructure | Retry after a short delay; check [status.openai.com](https://status.openai.com) |

**Example retry logic:**
```python
import time, openai

for attempt in range(5):
    try:
        response = client.chat.completions.create(...)
        break
    except openai.RateLimitError:
        wait = 2 ** attempt          # 1s, 2s, 4s, 8s, 16s
        print(f"Rate limited, retrying in {wait}s...")
        time.sleep(wait)
```

---

## Rate Limits

OpenAI rate limits are enforced per API key and vary by usage tier. Limits apply in two dimensions:

| Dimension | Description |
|---|---|
| **RPM** (Requests per minute) | Maximum API calls within a 60-second window |
| **TPM** (Tokens per minute) | Maximum tokens (input + output) within a 60-second window |

**Default Tier 1 limits (approximate):**

| Model | RPM | TPM |
|---|---|---|
| GPT-4o | 500 | 30,000 |
| GPT-3.5-turbo | 3,500 | 60,000 |
| text-embedding-3-small | 3,000 | 1,000,000 |
| DALL-E 3 | 5 images/min | — |

> Limits increase automatically as your usage history grows. Check your current limits at [platform.openai.com/account/limits](https://platform.openai.com/account/limits).

**Best practices:**
- Batch embedding requests (up to 2048 inputs per call)
- Use `stream: true` for long completions to improve perceived latency
- Cache embeddings to avoid re-computing identical inputs
- Monitor token usage via the `usage` field in every response

---

## Further Reading

- [OpenAI API Reference](https://platform.openai.com/docs/api-reference)
- [Chat Completions Guide](https://platform.openai.com/docs/guides/chat)
- [Embeddings Guide](https://platform.openai.com/docs/guides/embeddings)
- [Image Generation Guide](https://platform.openai.com/docs/guides/images)
- [Rate Limits](https://platform.openai.com/docs/guides/rate-limits)
- [Pricing](https://openai.com/pricing)
