# ============================================================================
# API GATEWAY ROUTES - Point d'entrée unique (Port 8080)
# ============================================================================

# Routes d'administration
GET    /health                    # Health check global
GET    /metrics                   # Métriques Prometheus  
GET    /api/version               # Version de l'API
GET    /api/status                # Statut des services

# Routes d'authentification (proxy vers Auth Service)
POST   /api/v1/auth/login         # Connexion utilisateur
POST   /api/v1/auth/logout        # Déconnexion
POST   /api/v1/auth/refresh       # Refresh token JWT
POST   /api/v1/auth/register      # Inscription (si activée)
POST   /api/v1/auth/forgot        # Mot de passe oublié
POST   /api/v1/auth/reset         # Reset mot de passe
GET    /api/v1/auth/me            # Profil utilisateur actuel

# Routes agrégées (combinant plusieurs services)
GET    /api/v1/dashboard          # Dashboard avec stats de tous les services
GET    /api/v1/search             # Recherche globale (properties + tenants)
GET    /api/v1/overview          # Vue d'ensemble (propriétés + locataires + finances)

# Proxy vers Property Service
GET    /api/v1/properties         # Liste des propriétés
POST   /api/v1/properties         # Créer propriété
GET    /api/v1/properties/{id}    # Détail propriété
PUT    /api/v1/properties/{id}    # Modifier propriété
DELETE /api/v1/properties/{id}    # Supprimer propriété

# Proxy vers Tenant Service  
GET    /api/v1/tenants           # Liste des locataires
POST   /api/v1/tenants           # Créer locataire
GET    /api/v1/tenants/{id}      # Détail locataire
PUT    /api/v1/tenants/{id}      # Modifier locataire
DELETE /api/v1/tenants/{id}      # Supprimer locataire

# Proxy vers Payment Service
GET    /api/v1/payments          # Liste des paiements
POST   /api/v1/payments          # Créer paiement
GET    /api/v1/payments/{id}     # Détail paiement

# Proxy vers Contract Service
GET    /api/v1/contracts         # Liste des contrats
POST   /api/v1/contracts         # Créer contrat
GET    /api/v1/contracts/{id}    # Détail contrat

# Proxy vers Notification Service
GET    /api/v1/notifications     # Liste notifications
POST   /api/v1/notifications     # Envoyer notification

# ============================================================================
# AUTH SERVICE - Authentification et autorisation (Port 8081)
# ============================================================================

# Health et monitoring
GET    /health                    # Health check du service
GET    /metrics                   # Métriques spécifiques à l'auth
GET    /ready                     # Readiness check

# Authentification
POST   /auth/login               # Connexion utilisateur
POST   /auth/logout              # Déconnexion (invalide le token)
POST   /auth/refresh             # Refresh du token JWT
POST   /auth/verify              # Vérification token (pour autres services)

# Gestion des utilisateurs
GET    /users                    # Liste des utilisateurs (admin only)
POST   /users                    # Créer un utilisateur
GET    /users/{id}               # Profil d'un utilisateur
PUT    /users/{id}               # Modifier un utilisateur
DELETE /users/{id}               # Supprimer un utilisateur
GET    /users/me                 # Profil de l'utilisateur connecté
PUT    /users/me                 # Modifier son propre profil
PUT    /users/me/password        # Changer son mot de passe

# Gestion des rôles et permissions
GET    /roles                    # Liste des rôles disponibles
POST   /roles                    # Créer un rôle (admin only)
GET    /roles/{id}               # Détail d'un rôle
PUT    /roles/{id}               # Modifier un rôle
DELETE /roles/{id}               # Supprimer un rôle

GET    /permissions              # Liste des permissions
POST   /users/{id}/roles         # Assigner un rôle à un utilisateur
DELETE /users/{id}/roles/{roleId} # Retirer un rôle

# Gestion des tenants (multi-tenant)
GET    /tenants                  # Liste des tenants (superadmin only)
POST   /tenants                  # Créer un tenant
GET    /tenants/{id}             # Détail d'un tenant
PUT    /tenants/{id}             # Modifier un tenant
DELETE /tenants/{id}             # Supprimer un tenant
GET    /tenants/{id}/users       # Utilisateurs d'un tenant

# Sessions et sécurité
GET    /sessions                 # Sessions actives de l'utilisateur
DELETE /sessions/{id}            # Fermer une session spécifique
DELETE /sessions/all             # Fermer toutes les sessions
GET    /audit-logs               # Logs d'audit (admin only)

# Mot de passe et sécurité
POST   /password/forgot          # Demande de reset mot de passe
POST   /password/reset           # Reset mot de passe avec token
POST   /password/validate        # Vérifier la force d'un mot de passe

# ============================================================================
# PROPERTY SERVICE - Gestion des biens immobiliers (Port 8082)
# ============================================================================

# Health et monitoring
GET    /health                   # Health check
GET    /metrics                  # Métriques du service
GET    /ready                    # Readiness check

# CRUD des propriétés
GET    /properties               # Liste des propriétés avec filtres
POST   /properties               # Créer une nouvelle propriété
GET    /properties/{id}          # Détail d'une propriété
PUT    /properties/{id}          # Modifier une propriété
DELETE /properties/{id}          # Supprimer une propriété
PATCH  /properties/{id}          # Modification partielle

# Recherche et filtres
GET    /properties/search        # Recherche textuelle avancée
GET    /properties/filter        # Filtres avancés (prix, type, etc.)
GET    /properties/nearby        # Propriétés à proximité (géolocalisation)
GET    /properties/available     # Propriétés disponibles seulement
GET    /properties/rented        # Propriétés louées

# Gestion des médias
POST   /properties/{id}/photos   # Ajouter des photos
GET    /properties/{id}/photos   # 

DELETE /properties/{id}/photos/{photoId} # Supprimer une photo
PUT    /properties/{id}/photos/{photoId}  # Modifier une photo

POST   /properties/{id}/documents # Ajouter des documents
GET    /properties/{id}/documents # Liste des documents
DELETE /properties/{id}/documents/{docId} # Supprimer un document

# Visites et disponibilité
GET    /properties/{id}/visits   # Planning des visites
POST   /properties/{id}/visits   # Programmer une visite
PUT    /properties/{id}/visits/{visitId} # Modifier une visite
DELETE /properties/{id}/visits/{visitId} # Annuler une visite

PUT    /properties/{id}/status   # Changer le statut (disponible/loué/maintenance)
GET    /properties/{id}/availability # Disponibilité et planning

# Statistiques et rapports
GET    /properties/stats         # Statistiques générales
GET    /properties/{id}/history  # Historique d'une propriété
GET    /properties/reports       # Rapports (occupation, revenus, etc.)

# Types et catégories
GET    /property-types           # Types de propriétés disponibles
GET    /property-features        # Caractéristiques disponibles
GET    /locations               # Localisations/quartiers

# Gestion des favoris
POST   /properties/{id}/favorite # Ajouter aux favoris
DELETE /properties/{id}/favorite # Retirer des favoris
GET    /favorites               # Liste des propriétés favorites

# ============================================================================
# TENANT SERVICE - Gestion des locataires (Port 8083)
# ============================================================================

# Health et monitoring
GET    /health                   # Health check
GET    /metrics                  # Métriques du service
GET    /ready                    # Readiness check

# CRUD des locataires
GET    /tenants                  # Liste des locataires
POST   /tenants                  # Créer un nouveau locataire
GET    /tenants/{id}             # Profil complet d'un locataire
PUT    /tenants/{id}             # Modifier un locataire
DELETE /tenants/{id}             # Supprimer un locataire
PATCH  /tenants/{id}             # Modification partielle

# Recherche et filtres
GET    /tenants/search           # Recherche par nom, email, etc.
GET    /tenants/filter           # Filtres (statut, revenus, etc.)
GET    /tenants/active           # Locataires actuels
GET    /tenants/candidates       # Candidats locataires
GET    /tenants/former           # Anciens locataires

# Gestion des documents
POST   /tenants/{id}/documents   # Ajouter des documents
GET    /tenants/{id}/documents   # Liste des documents
PUT    /tenants/{id}/documents/{docId} # Modifier un document
DELETE /tenants/{id}/documents/{docId} # Supprimer un document
GET    /tenants/{id}/documents/{docId}/download # Télécharger un document

# Dossier de candidature
POST   /tenants/{id}/application # Créer une candidature
GET    /tenants/{id}/application # Détail de la candidature
PUT    /tenants/{id}/application # Modifier la candidature
GET    /tenants/{id}/application/status # Statut de la candidature

# Garanties et cautions
POST   /tenants/{id}/guarantors  # Ajouter un garant
GET    /tenants/{id}/guarantors  # Liste des garants
PUT    /tenants/{id}/guarantors/{guarantorId} # Modifier un garant
DELETE /tenants/{id}/guarantors/{guarantorId} # Supprimer un garant

# Historique et références
GET    /tenants/{id}/history     # Historique de location
POST   /tenants/{id}/references  # Ajouter une référence
GET    /tenants/{id}/references  # Liste des références
DELETE /tenants/{id}/references/{refId} # Supprimer une référence

# Communications
GET    /tenants/{id}/messages    # Historique des messages
POST   /tenants/{id}/messages    # Envoyer un message
GET    /tenants/{id}/appointments # Rendez-vous programmés
POST   /tenants/{id}/appointments # Programmer un rendez-vous

# Évaluations et scoring
GET    /tenants/{id}/score       # Score de solvabilité
PUT    /tenants/{id}/evaluation  # Évaluation du dossier
GET    /tenants/{id}/credit-check # Vérification de crédit

# Statistiques
GET    /tenants/stats            # Statistiques générales
GET    /tenants/reports          # Rapports sur les locataires

# ============================================================================
# PAYMENT SERVICE - Gestion des paiements (Port 8084)
# ============================================================================

# Health et monitoring
GET    /health                   # Health check
GET    /metrics                  # Métriques du service
GET    /ready                    # Readiness check

# CRUD des paiements
GET    /payments                 # Liste des paiements
POST   /payments                 # Créer/traiter un paiement
GET    /payments/{id}            # Détail d'un paiement
PUT    /payments/{id}            # Modifier un paiement
DELETE /payments/{id}            # Annuler un paiement (si possible)

# Types de paiements
GET    /payments/rent            # Paiements de loyer
GET    /payments/deposits        # Dépôts de garantie
GET    /payments/charges         # Charges et frais
GET    /payments/late-fees       # Pénalités de retard

# Gestion des loyers
POST   /rent/schedule            # Programmer les loyers récurrents
GET    /rent/{tenantId}/schedule # Planning des loyers d'un locataire
PUT    /rent/{scheduleId}        # Modifier un planning de loyer
GET    /rent/overdue             # Loyers en retard

# Factures et reçus
GET    /invoices                 # Liste des factures
POST   /invoices                 # Créer une facture
GET    /invoices/{id}            # Détail d'une facture
PUT    /invoices/{id}            # Modifier une facture
POST   /invoices/{id}/send       # Envoyer une facture

GET    /receipts                 # Liste des reçus
GET    /receipts/{paymentId}     # Reçu d'un paiement
POST   /receipts/{paymentId}/send # Envoyer un reçu

# Intégrations de paiement
POST   /stripe/webhook           # Webhook Stripe
POST   /paypal/webhook           # Webhook PayPal
GET    /payment-methods          # Méthodes de paiement disponibles
POST   /payment-methods          # Ajouter une méthode de paiement

# Relances et rappels
GET    /reminders                # Liste des rappels
POST   /reminders                # Créer un rappel de paiement
PUT    /reminders/{id}           # Modifier un rappel
DELETE /reminders/{id}           # Supprimer un rappel
POST   /reminders/{id}/send      # Envoyer un rappel

# Rapports financiers
GET    /reports/income           # Rapport des revenus
GET    /reports/expenses         # Rapport des dépenses
GET    /reports/cash-flow        # Flux de trésorerie
GET    /reports/tax              # Rapport fiscal
GET    /reports/rent-roll        # État des loyers

# Statistiques
GET    /stats/payments           # Statistiques des paiements
GET    /stats/revenue            # Statistiques de revenus
GET    /stats/defaults           # Statistiques des impayés

# Comptabilité
GET    /accounts                 # Comptes comptables
POST   /transactions             # Créer une écriture comptable
GET    /transactions             # Liste des écritures
GET    /balance-sheet            # Bilan comptable

# ============================================================================
# CONTRACT SERVICE - Gestion des contrats et baux (Port 8085)
# ============================================================================

# Health et monitoring
GET    /health                   # Health check
GET    /metrics                  # Métriques du service
GET    /ready                    # Readiness check

# CRUD des contrats
GET    /contracts                # Liste des contrats
POST   /contracts                # Créer un nouveau contrat
GET    /contracts/{id}           # Détail d'un contrat
PUT    /contracts/{id}           # Modifier un contrat
DELETE /contracts/{id}           # Supprimer un contrat (brouillon)

# Gestion du cycle de vie
POST   /contracts/{id}/draft     # Créer un brouillon
PUT    /contracts/{id}/finalize  # Finaliser le contrat
POST   /contracts/{id}/sign      # Signer le contrat
POST   /contracts/{id}/activate  # Activer le contrat
POST   /contracts/{id}/terminate # Résilier le contrat

# Templates et modèles
GET    /templates                # Liste des modèles de contrat
POST   /templates                # Créer un modèle
GET    /templates/{id}           # Détail d'un modèle
PUT    /templates/{id}           # Modifier un modèle
DELETE /templates/{id}           # Supprimer un modèle

# Signatures électroniques
POST   /contracts/{id}/signature-request # Demander signature
GET    /contracts/{id}/signatures # Statut des signatures
POST   /contracts/{id}/signatures/{signatureId}/sign # Signer
GET    /signature/{token}        # Page de signature publique

# Renouvellements
GET    /contracts/{id}/renewal   # Détail du renouvellement
POST   /contracts/{id}/renewal   # Proposer un renouvellement
PUT    /contracts/{id}/renewal   # Modifier les conditions de renouvellement
POST   /contracts/{id}/renewal/accept # Accepter le renouvellement

# Documents et annexes
POST   /contracts/{id}/documents # Ajouter des documents
GET    /contracts/{id}/documents # Liste des documents
DELETE /contracts/{id}/documents/{docId} # Supprimer un document
GET    /contracts/{id}/pdf       # Générer PDF du contrat

# État des lieux
POST   /contracts/{id}/inventory/entry # État des lieux d'entrée
GET    /contracts/{id}/inventory/entry # Consulter EDL d'entrée
POST   /contracts/{id}/inventory/exit  # État des lieux de sortie
GET    /contracts/{id}/inventory/exit  # Consulter EDL de sortie

# Clauses et conditions
GET    /clauses                  # Liste des clauses disponibles
POST   /clauses                  # Créer une clause
GET    /contracts/{id}/clauses   # Clauses d'un contrat
POST   /contracts/{id}/clauses   # Ajouter une clause
DELETE /contracts/{id}/clauses/{clauseId} # Supprimer une clause

# Alertes et échéances
GET    /contracts/expiring       # Contrats arrivant à échéance
GET    /contracts/expired        # Contrats expirés
GET    /alerts                   # Alertes sur les contrats

# Rapports
GET    /reports/contracts        # Rapport sur les contrats
GET    /reports/occupancy        # Taux d'occupation
GET    /stats/contracts          # Statistiques des contrats

# ============================================================================
# NOTIFICATION SERVICE - Gestion des notifications (Port 8086)
# ============================================================================

# Health et monitoring
GET    /health                   # Health check
GET    /metrics                  # Métriques du service
GET    /ready                    # Readiness check

# CRUD des notifications
GET    /notifications            # Liste des notifications
POST   /notifications            # Créer/envoyer une notification
GET    /notifications/{id}       # Détail d'une notification
PUT    /notifications/{id}       # Modifier une notification
DELETE /notifications/{id}       # Supprimer une notification

# Gestion par canal
POST   /email                    # Envoyer un email
POST   /sms                      # Envoyer un SMS
POST   /push                     # Envoyer notification push
POST   /whatsapp                 # Envoyer message WhatsApp (si intégré)

# Templates de messages
GET    /templates                # Liste des templates
POST   /templates                # Créer un template
GET    /templates/{id}           # Détail d'un template
PUT    /templates/{id}           # Modifier un template
DELETE /templates/{id}           # Supprimer un template

# Notifications programmées
POST   /scheduled                # Programmer une notification
GET    /scheduled                # Liste des notifications programmées
PUT    /scheduled/{id}           # Modifier une notification programmée
DELETE /scheduled/{id}           # Annuler une notification programmée

# Notifications automatiques
GET    /automation/rules         # Règles d'automatisation
POST   /automation/rules         # Créer une règle
PUT    /automation/rules/{id}    # Modifier une règle
DELETE /automation/rules/{id}    # Supprimer une règle

# Préférences utilisateur
GET    /preferences/{userId}     # Préférences d'un utilisateur
PUT    /preferences/{userId}     # Modifier les préférences
POST   /unsubscribe/{token}      # Désabonnement (lien email)

# Historique et tracking
GET    /history                  # Historique des envois
GET    /history/{id}/status      # Statut d'envoi d'une notification
GET    /analytics/delivery       # Analytics de délivrance
GET    /analytics/engagement     # Analytics d'engagement

# Webhooks entrants
POST   /webhooks/sendgrid        # Webhook SendGrid
POST   /webhooks/twilio          # Webhook Twilio
POST   /webhooks/mailgun         # Webhook Mailgun

# ============================================================================
# ROUTES AGRÉGÉES (API Gateway uniquement)
# ============================================================================

# Dashboard et vues d'ensemble
GET    /api/v1/dashboard/owner           # Dashboard propriétaire
GET    /api/v1/dashboard/manager         # Dashboard gestionnaire
GET    /api/v1/dashboard/tenant          # Dashboard locataire

# Recherche globale
GET    /api/v1/search?q={query}          # Recherche dans properties + tenants
GET    /api/v1/search/properties?q={query} # Recherche properties uniquement
GET    /api/v1/search/tenants?q={query}    # Recherche tenants uniquement

# Rapports combinés
GET    /api/v1/reports/complete          # Rapport complet (tous services)
GET    /api/v1/reports/financial         # Rapport financier agrégé
GET    /api/v1/reports/occupancy         # Rapport d'occupation global

# Workflows complexes
POST   /api/v1/workflows/new-tenant      # Workflow complet nouveau locataire
POST   /api/v1/workflows/rent-payment    # Workflow paiement loyer
POST   /api/v1/workflows/property-visit  # Workflow visite propriété

# ============================================================================
# PARAMÈTRES COMMUNS POUR TOUTES LES ROUTES
# ============================================================================

# Headers requis
Authorization: Bearer {jwt_token}    # Token JWT obligatoire (sauf auth)
X-Tenant-ID: {tenant_uuid}           # ID du tenant (multi-tenant)
Content-Type: application/json       # Type de contenu
Accept: application/json             # Type de réponse souhaité

# Paramètres de pagination (pour les listes)
?page=1                              # Numéro de page (défaut: 1)
?limit=20                            # Nombre d'éléments par page (défaut: 20)
?sort=created_at                     # Champ de tri
?order=desc                          # Ordre de tri (asc/desc)

# Paramètres de filtrage communs
?search={query}                      # Recherche textuelle
?status={status}                     # Filtrer par statut
?created_after={date}                # Créé après une date
?created_before={date}               # Créé avant une date
?updated_after={date}                # Modifié après une date

# Paramètres d'inclusion (pour optimiser les requêtes)
?include=photos,documents            # Inclure des relations
?fields=id,name,status              # Sélectionner des champs spécifiques

# ============================================================================
# CODES DE RÉPONSE HTTP STANDARDS
# ============================================================================

# Succès
200 OK                               # Requête réussie
201 Created                          # Ressource créée
202 Accepted                         # Requête acceptée (traitement async)
204 No Content                       # Suppression réussie

# Erreurs client
400 Bad Request                      # Données invalides
401 Unauthorized                     # Non authentifié
403 Forbidden                        # Non autorisé
404 Not Found                        # Ressource introuvable
409 Conflict                         # Conflit (ex: email déjà utilisé)
422 Unprocessable Entity             # Données valides mais non traitables
429 Too Many Requests                # Rate limit dépassé

# Erreurs serveur
500 Internal Server Error            # Erreur serveur
502 Bad Gateway                      # Service indisponible
503 Service Unavailable              # Service temporairement indisponible
504 Gateway Timeout                  # Timeout du service