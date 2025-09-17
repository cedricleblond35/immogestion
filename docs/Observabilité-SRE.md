# Observabilité & SRE
Instrumentation : OpenTelemetry SDK (Go/Python/Angular) → traces + métriques + logs corrélés.
Stack : Prometheus (metrics), Loki (logs), Tempo/Jaeger (tracing), Grafana (dashboards).
SLOs (ex.) :
    API p95 latency < 150 ms (interne), < 300 ms (externe)
    Uptime 99.9%/mois pour Gateway
    Event lag Kafka < 2 s p99
Alerting : Alertmanager (CPU/mem, error rate, queue lag, 5xx, budget error SLO).
Chaos/DR : chaos testing (litmus/chaos-mesh), RPO 15 min (WAL shipping), RTO 1 h.

# CI/CD & Infrastructure
Repos
    Mono-repo immogestion/ avec dossiers /services/<name> + /libs, ou polyrepo si tu préfères.
    Contrats d’API et schémas événements versionnés dans /contracts/ (OpenAPI/Proto).

CI (GitHub Actions)
    lint + tests + SAST (gosec/bandit)
    build images (Docker) + SBOM (Syft) + vuln scan (Trivy/Grype)
    contract testing (Pact) + fixtures e2e (kind)

CD
    Helm charts par service + Argo CD (GitOps, PR-based envs)
    envs : dev (ephemeral preview), staging, prod

Kubernetes
    Istio (mTLS, canary/traffic shifting)
    HPA/VPA, Pod Disruption Budgets, PodSecurity, ResourceQuotas
    Node pools dédiés (généraux, mémoire, GPU pour AI)

Stateful
    Postgres (managed : CloudSQL/Aurora) + Patroni si self-host, read-replicas
    MongoDB Atlas, Kafka (Confluent/Redpanda), Redis (managed)

# Résilience & patterns critiques
Outbox Pattern : chaque service qui émet des events écrit d’abord dans sa DB (tx locale) puis CDC (Debezium) publie vers Kafka → exactly-once effectif côté consommateur via idempotency keys.

Sagas : orchestrées par événements (ex. création projet → provisioning participants → notifications).

Idempotence : Idempotency-Key sur endpoints write, compaction Kafka par clé si pertinent.

Backpressure : quotas consumers, pause/resume, DLQ par topic.

Feature flags : Unleash/Flagsmith pour déploiements progressifs.

# Multitenant (B2B)

Modèle : single-DB shared schema + org_id partout (plus simple) ou sharding par org pour gros clients.

Isolation :

RLS Postgres (Row-Level Security) sur org_id

Buckets S3 par org/projet

Kafka: partitions par org_id si besoin de débit.

Billing : events d’usage (stockage, analyses IA, utilisateurs actifs) → billing.usage.v1.

# Plan de tests (qualité en continu)

Unit : 80%+ sur services critiques (Auth, Projects, Docs).

Contract tests : Pact (consumers/producers REST & gRPC), schema registry tests (Kafka).

Integration : docker-compose/kind, tests e2e (Playwright/Cypress pour Angular).

Performance : k6 (APIs), kafka-bench (débit/latence), chaos (latence, kill pods).

Sécurité : SAST/DAST, fuzzing gRPC (bois).

Data : migrations (golang-migrate), seeding anonymisé, tests RLS.

# Services « concurrents » indispensables + services innovants (tes différenciateurs)

Indispensables (parité marché)

Devis/facturation (intégration compta), planning/Gantt, documents versionnés, rôles/permissions, chat/notifications, budget vs coût, mobile capture photo/signature.

Différenciateurs innovants

Vision chantier : % avancement, détection sécurité (EPI), heatmap progrès.

Détection risque retard/surcoût (ML) : alerte proactive + recommandations.

Sourcing intelligent (achats) : alternatives fournisseur si rupture, suggestions prix.

Mode offline-first solide (sync confluent) pour le terrain.

API-first + plugins : store d’extensions (connecteurs, widgets métier).

Marketplace prestataires (commissions), onboarding KYC/assurance intégrés.

Audit & conformité augmentés : génération semi-automatique de livrables (DOE, PPSPS…).