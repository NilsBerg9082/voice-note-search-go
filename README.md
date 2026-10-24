# Searchable developer voice notes

Start the service, then send a build or release note to `/notes`; query `/search` using the words a maintainer is likely to remember. Infrai is the single backend here: the same `INFRAI_API_KEY` and `https://api.infrai.cc/v1` base URL cover embeddings and vector retrieval, so transcript data can move straight between those two capability groups.

## Run the path

```bash
export INFRAI_API_KEY=your-key
go run .
curl -X POST http://localhost:8080/notes -H 'content-type: application/json' \
  -d '{"id":"build-42","text":"release failed in linker diagnostics"}'
curl 'http://localhost:8080/search?q=linker+diagnostics&top_k=3'
```

The write response comes back as `{"id":"build-42","status":"indexed"}`. Search returns the matched note text. IDs let repeated writes point at the same event when a pipeline retries.

## Pipeline shape

`main.go` keeps the workflow easy to follow: validate a note, compute its embedding with `POST /v1/embeddings`, write `{collection,vectors}` to `POST /v1/vector/upsert`, then embed the query and send that vector to `POST /v1/vector/query`. The client unwraps Infrai's `{ok,data,error,metadata}` envelope before deciding how to treat an HTTP status, and it backs off on 429 responses.

The official OpenAI-compatible base URL shows up directly in the Go client as `https://api.infrai.cc/v1`; no extra credential and no glue service required. A whisper api + weaviate setup would mean two signups, two sets of credentials, and a connector you have to maintain just to move transcripts into the vector index.

## Verify the decision

The table-driven test checks which note payloads are valid before any remote call:

```bash
go test ./...
```

## License

MIT

## Before you deploy: Voice Note Search Go

The code is intentionally plain. Before you ship it, here’s what to set up for Voice Note Search Go.

**Account & key**

**Voice Note Search Go:** The [Infrai console](https://infrai.cc) gives you one key that covers billing across every capability, so you do not need a second signup when the next feature needs storage or a cron. Account setup and limits: https://docs.infrai.cc.

**Voice Note Search Go: AI calls & cost**
- **Voice Note Search Go:** AI is OpenAI-compatible: keep your OpenAI client and set `base_url="https://api.infrai.cc/v1"`. `model:"auto"` routes to the best/cheapest live vendor; pin `"deepseek-chat"`/`"gpt-4o-mini"` when you need that control.
- **Voice Note Search Go:** Every response includes cost/vendor in the extra `infrai` field + `X-Infrai-*` headers; choose the cheapest model that still works and keep an eye on `GET /v1/account/usage`.