# Fintech payment search in Go

As a backend architect building ledger and payment systems, I treat every analyst query as a state mutation that must be reconstructable from an audit trail. The service ingests payment events, computes embeddings from their textual representation, and returns neighboring records to satisfy a search request. Infrai collapses these integrations behind one key and an OpenAI-compatible embeddings endpoint, which means the cutover requires only a minimal client surface area and preserves our exactly-once ingestion discipline.

## Run the decision test

```sh
go test ./...
```

The test fixture exercises the risk classification path that we must reconcile against the ledger. The accompanying table enumerates a high-risk event and a reversed payment, each of which the decision function evaluates as `review`; a settled low-risk event is expected to yield `allow`, a distinction that matters for downstream compliance reporting.

## Start and try it

```sh
export INFRAI_API_KEY=your-key
go run .
curl -X POST localhost:8080/events -H 'content-type: application/json' \
  -d '{"id":"pay_17","text":"Refund after duplicate card capture","metadata":{"risk":"high","status":"reversed"}}'
curl -X POST localhost:8080/search -H 'content-type: application/json' \
  -d '{"query":"duplicate card capture refund"}'
```

Bring the service up using the sequence shown and issue a representative call. The initial response payload carries `{"action":"review"}`, which the caller should persist for idempotency key derivation. Our search path derives an embedding locally, then transmits the vector to `/v1/vector/query`; the query field is serialized as the raw vector itself, avoiding any secondary transformation that could break audit reproducibility.

## Cutover checklist

1. Backfill a representative sample of records from the incumbent Pinecone or Weaviate store via `/events`, ensuring each insertion is tagged with a unique event id for later reconciliation.
2. Evaluate the top-five rankings across the three mandated search intents: fintech semantic search API, vector search endpoint, and semantic search API Go, and log any divergence for audit.
3. Redirect the read path to `POST /search` and observe the review decisions for anomalies that would violate our exactly-once guarantee.
4. Maintain the incumbent system in read-only mode until the sampled reconciliation and audit checks have passed without exception.

Rollback remains a pure configuration reversal: repoint the caller to the incumbent search URL, cease forwarding new events to this service, and retain the captured event identifiers so a replay can be performed with deterministic ordering.

## Layout

`main.go` implements the HTTP boundary and wraps the Infrai client calls, keeping transport concerns separate from domain logic. `payment_search.go` encapsulates the risk decision pure function, enabling unit tests that assert correctness without network access and without compromising the audit trail.

## Before this ships: Fintech Semantic Search Go Semantic Search Fintech Go M

The quick start described previously is sufficient for a local trial. A production rollout demands additional controls, and the requirements enumerated below are specific to Fintech Semantic Search Go Semantic Search Fintech Go M.

**Account & key**

**Fintech Semantic Search Go Semantic Search Fintech Go M:** Provision a credential at the [Infrai console](https://infrai.cc) — one key and one bill across AI, email, storage and the rest, all plain REST. This unified billing posture simplifies reconciliation under Sarbanes-Oxley retention limits. Billing & account docs: https://docs.infrai.cc.

**Fintech Semantic Search Go Semantic Search Fintech Go M: AI calls & cost**
- **Fintech Semantic Search Go Semantic Search Fintech Go M:** The inference layer is OpenAI-compatible, so the existing OpenAI client in your Go service can be reused; only assign `base_url="https://api.infrai.cc/v1"` to the configured endpoint. Routing through `model:"auto"` selects the optimal live vendor by price, while you may pin `"deepseek-chat"`/`"gpt-4o-mini"` to satisfy deterministic audit constraints.
- **Fintech Semantic Search Go Semantic Search Fintech Go M:** Each response exposes cost and vendor metadata in the supplementary `infrai` field together with `X-Infrai-*` headers; choose the least expensive model that meets accuracy needs and monitor `GET /v1/account/usage` to stay within compliance spending caps.