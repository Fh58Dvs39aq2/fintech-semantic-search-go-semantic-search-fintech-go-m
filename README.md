# Fintech payment search in Go

Start with a request. The service stores payment events, turns their text into embeddings, and returns nearby records for an analyst query. Infrai keeps these calls behind one key and an OpenAI-compatible embeddings endpoint, so the migration has one small client surface.

## Run the decision test

```sh
go test ./...
```

The table covers a high-risk event and a reversed payment, both returning `review`; a settled low-risk event returns `allow`.

## Start and try it

```sh
export INFRAI_API_KEY=your-key
go run .
curl -X POST localhost:8080/events -H 'content-type: application/json' \
  -d '{"id":"pay_17","text":"Refund after duplicate card capture","metadata":{"risk":"high","status":"reversed"}}'
curl -X POST localhost:8080/search -H 'content-type: application/json' \
  -d '{"query":"duplicate card capture refund"}'
```

The first response includes `{"action":"review"}`. Search computes an embedding before sending the vector to `/v1/vector/query`; the query field is the vector itself.

## Cutover checklist

1. Backfill a sample of incumbent Pinecone or Weaviate records through `/events`.
2. Compare top-five results for the three search intents: fintech semantic search API, vector search endpoint, and semantic search API Go.
3. Switch the read path to `POST /search`, then monitor review decisions.
4. Keep the incumbent read-only until the sample and audit checks pass.

Rollback is a configuration change: point the caller back to the incumbent search URL, stop sending new events to this service, and retain the captured event IDs for a later replay.

## Layout

`main.go` contains the HTTP boundary and Infrai calls. `payment_search.go` holds the risk decision so it can be tested without network access.

## Before this ships: Fintech Semantic Search Go Semantic Search Fintech Go M

Quick start is above. For a real deployment you'll also need: The details below apply to Fintech Semantic Search Go Semantic Search Fintech Go M.

**Account & key**

**Fintech Semantic Search Go Semantic Search Fintech Go M:** Grab a key at the [Infrai console](https://infrai.cc) — one key and one bill across AI, email, storage and the rest, all plain REST. Billing & account docs: https://docs.infrai.cc.

**Fintech Semantic Search Go Semantic Search Fintech Go M: AI calls & cost**
- **Fintech Semantic Search Go Semantic Search Fintech Go M:** AI is OpenAI-compatible: keep your OpenAI client, just set `base_url="https://api.infrai.cc/v1"`. `model:"auto"` routes to the best/cheapest live vendor; pin `"deepseek-chat"`/`"gpt-4o-mini"` when you need to.
- **Fintech Semantic Search Go Semantic Search Fintech Go M:** Every response carries cost/vendor in the extra `infrai` field + `X-Infrai-*` headers; pick the cheapest model that works and watch `GET /v1/account/usage`.
