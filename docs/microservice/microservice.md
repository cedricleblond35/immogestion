
[Angular Web/App] ── BFF (optional) ─┐
                                     │
                               [API Gateway]
                                     │
         ┌───────────── Sync (gRPC/REST) ──────────────┐
         ▼                                             ▼
 [AuthZ/AuthN Svc]                               [Projects Svc]
 [Identity/OIDC]                                 [Tasks/Planning]
         │                                             │
         └─────────────┐                       ┌────────┘
                       │                       │
                     [Kafka]  <── Event Bus ──┘
                       │
   ┌───────────────────┼──────────────────────────────────────────────┐
   ▼                   ▼                   ▼                          ▼
[Docs Svc]        [Comms Svc]        [Budget Svc]               [AI Svc]
(S3/MinIO+MDB)    (Chat/Notif)       (Costs/ROI)                (OCR/Vision)
   │                   │                 │                         │
   └─► MongoDB     Redis Streams     Postgres                 GPU Workers
       (metadata)    + Kafka         (finance)              (K8s node pool)
           │            │                │                         │
           └──────► Object Storage (S3/MinIO) ◄────────────────────┘

            [Integrations Svc]──(ERP/Compta/Email/SMS/IdP)
                   │
                Connectors
       (webhooks, polling, ETL via Kafka/Outbox)

Observabilité: OpenTelemetry → Prometheus + Loki + Tempo/Jaeger + Grafana
Sécurité: OIDC (Keycloak/Auth0), mTLS service-to-service (Istio), Vault/Secrets
Orchestration: Kubernetes + Helm/Argo CD, Horizontal/Vertical Pod Autoscaler


# Services & limites fonctionnelles (bounded contexts)

## 1. Auth & Identity Service (Go)

Rôle : identité, organisations/tenants, rôles & permissions, OAuth2/OIDC (Keycloak recommandé).

Interfaces :

REST publique : /oauth2, /users, /orgs

gRPC interne : Authorize(subject, action, resource)

Dépôts : Postgres (users, orgs, roles, memberships)

Événements : UserRegistered v1, OrgCreated v1

Notes : Policy-as-code (OPA/OPAL) possible pour autorisations complexes.

## 2. Projects Service (Go)

Rôle : projets/chantier, lots, jalons, dépendances, états.

REST : POST /projects, GET /projects/{id}, PATCH /projects/{id}

gRPC : GetProject, ListProjects, UpdateStatus

DB : Postgres (projects, milestones, participants)

Events (Kafka) : ProjectCreated v1, ProjectUpdated v1, ParticipantAdded v1

## 3. Tasks/Planning Service (Go)

Rôle : tâches (Kanban/Gantt), assignations, progression, dépendances.

REST : POST /projects/{id}/tasks, PATCH /tasks/{id}/status

DB : Postgres (tasks, task_dependencies, assignments)

Events : TaskCreated v1, TaskStatusChanged v1, TaskDelayed v1

Algo : calcul chemin critique (background worker).

## 4. Documents Service (Go + Python worker)

Rôle : stockage fichiers, versioning, métadonnées, OCR async.

REST : POST /documents, GET /documents/{id}, POST /documents/{id}/versions

DB : MongoDB (documents, versions, ocr_text) + Object Store S3/MinIO

Events : DocumentUploaded v1, DocumentOcrReady v1

Pipelines : Outbox (upload) → Kafka → AI Worker OCR → callback.

## 5. Communications Service (Go)

Rôle : chat projet, fils de discussion liés à doc/tâche, mentions, notifications.

Tech : Redis Streams pour chat temps réel + Kafka pour événements persistés.

REST : POST /messages, GET /projects/{id}/messages

Events : MessagePosted v1, MentionCreated v1, NotificationDispatched v1

## 6. Budget & Finance Service (Go)

Rôle : budget initial, coûts réels, variances, alertes dérives.

REST : POST /projects/{id}/budget, POST /expenses, GET /.../budget-status

DB : Postgres (budgets, expense_lines, vendors)

Events : ExpenseRecorded v1, BudgetOverrunAlert v1

## 7. AI Service (Python)

Rôle : OCR (Tesseract/Google Vision local), NER (SpaCy), Vision chantier (PyTorch).

Interfaces : REST interne (/ocr, /vision/progress, /vision/safety) + consumer Kafka.

Compute : node pool GPU, autoscaling.

Events : OcrCompleted v1, ProgressEstimated v1, SafetyIssueDetected v1

## 8. Integrations Service (Go)

Rôle : connecteurs ERP/Compta (Sage/EBP/QuickBooks), Email/SMS (Sendgrid/Twilio), IdP.

Patterns : Webhooks entrants, polling, transformations (Avro/Protobuf), Outbox vers Kafka.

## 9. Notifications Service (Go)

Rôle : fan-out multi-canal (in-app, email, SMS, webhook).

DB : Postgres (notification_prefs, deliveries)

Events : consomme *Alert v1, MessagePosted v1, publie NotificationDispatched v1.

## 10. Reporting/Analytics Service (Go/Python)

Rôle : vues agrégées (SLO, délais, coûts), exports, datasets anonymisés.

Data : réplique en lecture (read-replica) Postgres + parquet/S3 pour historiques lourds.


