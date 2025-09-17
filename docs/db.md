# Entrer ds le conteneur

sudo docker exec -it immogestion_postgres_dev bash

# se connecter à Postgres

psql -h localhost -U immobilier_user -d immobilier_prod

\dn
\dt auth.\*
\d auth.users

select \* from auth.users;



Un exemple de clé secrète JWT (JWT_SECRET) générée de manière aléatoire est une chaîne hexadécimale de 64 caractères, produite en utilisant 32 octets aléatoires. Par exemple, une clé générée avec la commande Node.js :
$ node -e "console.log(require('crypto').randomBytes(32).toString('hex'))"

# Postgres – projects
CREATE TABLE projects (
  id UUID PRIMARY KEY,
  org_id UUID NOT NULL,
  name TEXT NOT NULL,
  status TEXT NOT NULL CHECK (status IN ('planned','active','on_hold','completed','cancelled')),
  start_date DATE, end_date DATE,
  created_at TIMESTAMPTZ DEFAULT now(),
  updated_at TIMESTAMPTZ DEFAULT now()
);
CREATE INDEX idx_projects_org ON projects(org_id);

# Postgres – tasks
CREATE TABLE tasks (
  id UUID PRIMARY KEY,
  project_id UUID NOT NULL REFERENCES projects(id),
  title TEXT NOT NULL,
  status TEXT NOT NULL CHECK (status IN ('todo','in_progress','blocked','done')),
  assignee UUID,
  planned_start DATE, planned_end DATE,
  progress SMALLINT DEFAULT 0 CHECK (progress BETWEEN 0 AND 100),
  created_at TIMESTAMPTZ DEFAULT now(), updated_at TIMESTAMPTZ DEFAULT now()
);

# MongoDB – documents
{
  "_id": "ObjectId",
  "project_id": "UUID",
  "org_id": "UUID",
  "filename": "plan_R+1.pdf",
  "content_type": "application/pdf",
  "storage_key": "s3://bucket/org/uuid/xxx",
  "versions": [
    {"v":1,"uploaded_by":"UUID","uploaded_at":"2025-09-16T10:20Z","checksum":"..."},
    {"v":2,"uploaded_by":"UUID","uploaded_at":"..."}
  ],
  "tags": ["plan","structural"],
  "ocr": {"status":"ready","lang":"fr","text_ref":"s3://.../ocr.txt"}
}

# Event – TaskStatusChanged v1 (protobuf)
message TaskStatusChangedV1 {
  string event_id = 1;
  string task_id = 2;
  string project_id = 3;
  string org_id = 4;
  string old_status = 5;
  string new_status = 6;
  google.protobuf.Timestamp occurred_at = 7;
  string changed_by = 8;
  int32 progress = 9; // 0..100
  string correlation_id = 10; // for sagas
  string idempotency_key = 11;
}
