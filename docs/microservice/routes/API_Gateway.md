# ============================================================================
# API GATEWAY - ROUTES COMPLÈTES (Port 8080)
# Point d'entrée unique pour tous les microservices
# ============================================================================

# ============================================================================
# ROUTES D'ADMINISTRATION ET MONITORING
# ============================================================================
GET    /health                           # Health check global de l'API Gateway
GET    /health/detailed                  # Health check détaillé de tous les services
GET    /ready                            # Readiness check (prêt à recevoir du trafic)
GET    /metrics                          # Métriques Prometheus de l'API Gateway
GET    /version                          # Version de l'API Gateway et services
GET    /status                           # Statut en temps réel de tous les services
GET    /status/{service}                 # Statut d'un service spécifique

# ============================================================================
# ROUTES D'AUTHENTIFICATION (Proxy vers Auth Service)
# ============================================================================

# Authentification de base
POST   /api/v1/auth/login               # Connexion utilisateur
POST   /api/v1/auth/logout              # Déconnexion (invalidation token)
POST   /api/v1/auth/refresh             # Renouvellement token JWT
GET    /api/v1/auth/me                  # Profil utilisateur connecté
PUT    /api/v1/auth/me                  # Modifier son profil
PUT    /api/v1/auth/me/password         # Changer son mot de passe

# Gestion mot de passe
POST   /api/v1/auth/password/forgot     # Demande reset mot de passe
POST   /api/v1/auth/password/reset      # Reset avec token reçu par email
POST   /api/v1/auth/password/validate   # Vérifier force mot de passe

# Inscription (si activée)
POST   /api/v1/auth/register            # Inscription nouveau compte
POST   /api/v1/auth/verify              # Vérification email inscription
POST   /api/v1/auth/resend-verification # Renvoyer email de vérification

# Sessions multi-device
GET    /api/v1/auth/sessions            # Sessions actives utilisateur
DELETE /api/v1/auth/sessions/{id}       # Fermer une session spécifique
DELETE /api/v1/auth/sessions/all        # Fermer toutes les sessions