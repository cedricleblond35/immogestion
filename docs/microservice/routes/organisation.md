ORGANISATION DES ROUTES API GATEWAY
===================================

API Gateway sert de POINT D'ENTRÉE UNIQUE pour:
├── 6 microservices backend (auth, property, tenant, payment, contract, notification)
├── 180+ routes au total
├── Authentification centralisée
├── Rate limiting et cache
└── Orchestration de workflows complexes

STRUCTURE HIÉRARCHIQUE DES ROUTES
=================================

/health, /metrics, /admin/*                    #  Administration
└── Monitoring, configuration, maintenance

/api/v1/auth/*                                 # Authentification
└── Login, logout, sessions, gestion comptes

/api/v1/properties/*                          # Gestion immobilière
└── CRUD biens, photos, visites, disponibilité

/api/v1/tenants/*                             # Gestion locataires
└── Profils, candidatures, documents, scoring

/api/v1/payments/*                            # Finances
└── Loyers, factures, relances, comptabilité

/api/v1/contracts/*                           # Juridique
└── Baux, signatures, renouvellements, EDL

/api/v1/notifications/*                       # Communications
└── Emails, SMS, templates, préférences

/api/v1/dashboard/*                           # Vues agrégées
└── Combinaison de données multi-services

/api/v1/workflows/*                           # Processus métier
└── Orchestration transactions complexes

EXEMPLES D'UTILISATION PRATIQUES
================================

SCÉNARIO 1: Application mobile - Recherche d'appartement
----------------------------------------------------------------
Requête Frontend:
GET /api/v1/properties/search?q=paris&type=apartment&max_price=1000
Headers:
  Authorization: Bearer eyJhbGciOiJIUzI1NiIs...
  X-Tenant-ID: 12345678-1234-1234-1234-123456789abc

API Gateway:
1. ✅ Vérifie le JWT (valid, non-expiré, non-blacklist)
2. ✅ Vérifie rate limit (user456: 45/100 req/min)
3. ✅ Vérifie cache Redis: MISS
4. 🔄 Route vers Property Service
5. 💾 Met en cache la réponse (TTL: 5min)
6. 📊 Incrémente métriques Prometheus

Réponse:
{
  "data": [
    {
      "id": "prop-123",
      "name": "Appartement 3P - Marais",
      "price": 950,
      "photos": ["url1", "url2"],
      "available": true
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 42
  }
}

💡 SCÉNARIO 2: Workflow complet - Nouveau locataire
---------------------------------------------------
Requête Frontend:
POST /api/v1/workflows/new-tenant
{
  "property_id": "prop-123",
  "tenant_data": {
    "name": "Jean Dupont",
    "email": "jean@email.com",
    "phone": "+33123456789"
  },
  "documents": ["id_card", "pay_slip", "bank_statement"]
}

API Gateway orchestre:
1. 🔐 Auth Service → Validation permissions
2. 🏠 Property Service → Vérification disponibilité
3. 👥 Tenant Service → Création profil + dossier
4. 📄 Contract Service → Génération brouillon bail
5. 📧 Notification Service → Email confirmation
6. 💾 Redis → Stockage état workflow

Réponse:
{
  "workflow_id": "wf-789",
  "status": "in_progress",
  "steps": [
    {"name": "tenant_created", "status": "completed"},
    {"name": "documents_uploaded", "status": "pending"},
    {"name": "contract_draft", "status": "waiting"}
  ],
  "next_action": "upload_documents"
}

💡 SCÉNARIO 3: Dashboard propriétaire - Vue d'ensemble
------------------------------------------------------
Requête Frontend:
GET /api/v1/dashboard/owner

API Gateway agrège données de:
1. 🏠 Property Service → Mes 15 propriétés, 3 libres, 12 louées
2. 👥 Tenant Service → 12 locataires actifs, 2 candidatures
3. 💰 Payment Service → 10,500€ revenus mois, 1 retard
4. 📄 Contract Service → 2 baux expire dans 3 mois
5. 📧 Notification Service → 5 messages non lus

Réponse agrégée:
{
  "summary": {
    "total_properties": 15,
    "occupied_properties": 12,
    "monthly_income": 10500.00,
    "occupancy_rate": 80.0
  },
  "alerts": [
    {"type": "late_payment", "count": 1},
    {"type": "expiring_contracts", "count": 2}
  ],
  "recent_activity": [...],
  "financial_overview": {...}
}

💡 SCÉNARIO 4: Rate limiting en action
--------------------------------------
Utilisateur fait 101 requêtes en 1 minute:

Requête #101:
GET /api/v1/properties

API Gateway:
1. 📊 Redis INCR rate_limit:tenant123:user456:GET_properties
2. ❌ Valeur = 101 > limite 100
3. 🚫 Retourne HTTP 429 Too Many Requests

Réponse:
{
  "error": "Rate limit exceeded",
  "message": "Maximum 100 requests per minute",
  "retry_after": 45,
  "reset_at": "2024-01-15T14:31:00Z"
}

💡 SCÉNARIO 5: Circuit breaker protection
-----------------------------------------
Property Service est down (3 pannes successives):

GET /api/v1/properties

API Gateway:
1. 🔍 Redis GET circuit:property-service → "OPEN"
2. ⚡ Circuit ouvert = service indisponible
3. 🔄 Retourne réponse de fallback ou erreur contrôlée

Réponse:
{
  "error": "Service temporarily unavailable",
  "message": "Property service is experiencing issues",
  "fallback_data": {...}, // Données en cache si disponibles
  "retry_after": 300
}

FONCTIONNALITÉS AVANCÉES DE L'API GATEWAY
=========================================

🔄 ORCHESTRATION DE WORKFLOWS
├── Transactions distribuées (Saga pattern)
├── Rollback automatique en cas d'échec
├── État persisté dans Redis
└── Retry avec backoff exponentiel

📊 AGRÉGATION DE DONNÉES
├── Combinaison réponses de plusieurs services
├── Cache intelligent des données agrégées
├── Optimisation des appels parallèles
└── Déduplication des requêtes identiques

🛡️ SÉCURITÉ MULTICOUCHE
├── Validation JWT centralisée
├── Autorisations RBAC par endpoint
├── Rate limiting par utilisateur/tenant/IP
├── Protection CORS et headers sécurisés
├── Chiffrement en transit (TLS)
└── Audit trail complet

⚡ PERFORMANCE ET RÉSILIENCE
├── Cache Redis multicouche
├── Circuit breakers par service
├── Load balancing automatique
├── Retry avec backoff intelligent
├── Compression des réponses
├── Keep-alive connections
└── Health checks continus

📈 MONITORING ET OBSERVABILITÉ
├── Métriques Prometheus détaillées
├── Tracing distribué (request-id)
├── Logs structurés avec corrélation
├── Alertes automatiques
├── Dashboard temps réel
└── SLA monitoring

CONFIGURATION EXEMPLE API GATEWAY
=================================

# Rate Limiting
rate_limit:
  default: 1000/hour
  authenticated: 5000/hour
  endpoints:
    "POST /api/v1/payments": 100/hour
    "GET /api/v1/properties": 1000/hour

# Cache Strategy  
cache:
  default_ttl: 300s
  policies:
    "GET /api/v1/properties": 900s
    "GET /api/v1/dashboard": 60s
    "GET /api/v1/tenants/*/score": 3600s

# Circuit Breakers
circuit_breakers:
  failure_threshold: 5
  recovery_timeout: 30s
  half_open_max_calls: 3

# Load Balancing
load_balancing:
  strategy: round_robin
  health_check_interval: 10s
  unhealthy_threshold: 3

MÉTRIQUES EXPOSÉES PAR L'API GATEWAY
===================================

📊 Métriques de performance:
- gateway_requests_total{method, endpoint, status}
- gateway_request_duration_seconds{method, endpoint}
- gateway_cache_hits_total{endpoint}
- gateway_cache_misses_total{endpoint}

🛡️ Métriques de sécurité:
- gateway_auth_attempts_total{result}
- gateway_rate_limit_exceeded_total{endpoint}
- gateway_circuit_breaker_state{service}

🔄 Métriques de services:
- gateway_service_requests_total{service, method, status}
- gateway_service_response_time{service}
- gateway_service_availability{service}

💾 Métriques Redis:
- gateway_redis_operations_total{operation}
- gateway_redis_cache_hit_ratio
- gateway_redis_connection_pool_size

COMMANDES UTILES POUR DEBUG
===========================

# Vérifier rate limiting
curl -H "Authorization: Bearer $TOKEN" \
     -H "X-Tenant-ID: $TENANT" \
     -v http://localhost:8080/api/v1/properties

# Tester circuit breaker
curl -H "Authorization: Bearer $TOKEN" \
     http://localhost:8080/admin/circuit-breaker/property-service/status

# Vider cache spécifique
curl -X POST \
     -H "Authorization: Bearer $ADMIN_TOKEN" \
     http://localhost:8080/admin/cache/clear?pattern=properties:*

# Monitoring en temps réel
curl http://localhost:8080/metrics | grep gateway_

# Health check détaillé
curl http://localhost:8080/health/detailed

L'API Gateway est le CERVEAU de votre architecture microservices 🧠
Il coordonne, protège, optimise et supervise tous vos services !