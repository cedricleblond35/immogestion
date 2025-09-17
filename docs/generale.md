2️⃣ Spécifications professionnelles – Microservices
2.1 Authentification & Gestion des rôles

Objectif : Gérer utilisateurs, permissions, organisations.

Entrées : email, mot de passe, organisation, rôle.

Sorties : token JWT, refresh token.

Rôles de base : Maître d’ouvrage (client), Architecte, Entreprise générale, Sous-traitant.

API :

POST /auth/register

POST /auth/login

GET /users/{id}

PATCH /users/{id}/role

Base : Postgres (schéma users, roles, permissions).

Tech : Go + gRPC/REST, Redis pour session.


2.2 Stockage documentaire

Objectif : Centraliser et versionner tous les documents (plans, devis, photos).

Fonctionnalités :

Upload/download fichiers.

Versioning automatique.

Métadonnées (type doc, auteur, date, projet lié).

OCR (Python) → extraction texte / montants / dates.

API :

POST /documents (upload)

GET /documents/{id}

GET /projects/{id}/documents

POST /documents/{id}/versions

Base : MongoDB (fichiers & versions), Postgres (liens projet ↔ document).

Tech : Go + MinIO/S3, Python OCR.


2.3 Communication multi-acteurs

Objectif : Remplacer WhatsApp/email, garder tout l’historique lié aux projets.

Fonctionnalités :

Chat temps réel par projet.

Commentaires liés à documents/tâches.

Notifications push (mail/SMS/app).

API :

POST /messages

GET /projects/{id}/messages

GET /documents/{id}/comments

Tech : Kafka (events), Redis Streams (chat temps réel), Go microservice.


2.4 Gestion projet & tâches

Objectif : Structurer les chantiers (projets, lots, tâches, planning).

Fonctionnalités :

Création projet.

Planning basique (Kanban, jalons).

Assignation acteurs.

Suivi progression (%).

API :

POST /projects

GET /projects/{id}

POST /projects/{id}/tasks

PATCH /tasks/{id}/status

Base : Postgres (schéma projets, tâches, dépendances).

Tech : Go + Angular dashboard (Kanban).


2.5 Budget & suivi financier (phase 2)

Objectif : Permettre aux MOA/entreprises de voir ROI et marge en direct.

Fonctionnalités :

Budget initial vs coûts réels.

Alertes dérives.

Graphiques comparatifs.

API :

POST /projects/{id}/budget

POST /projects/{id}/expenses

GET /projects/{id}/budget-status

Tech : Go, Postgres, Angular (charts).


2.6 IA OCR documents (phase 2)

Objectif : Automatiser la saisie des infos de devis, contrats, factures.

Fonctionnalités :

Extraction montants, dates, fournisseurs.

Indexation recherche plein texte.

Tech : Python (Tesseract, SpaCy), API REST.

Stockage : MongoDB (texte OCR).

2.7 IA Vision chantier (phase différenciation)

Objectif : Automatiser suivi chantier via photos/vidéos.

Fonctionnalités :

Détection avancement (fondations, murs, finitions).

Détection anomalies (absence EPI, sécurité).

Tech : Python (OpenCV, TensorFlow/PyTorch).

API :

POST /projects/{id}/images (analyse automatique).

2.8 Marketplace prestataires (phase différenciation)

Objectif : Offrir un écosystème (fournisseurs, artisans, assureurs).

Fonctionnalités :

Mise en relation.

Gestion devis/contrats depuis la plateforme.

Commission sur transactions.

Base : Postgres (catalogue prestataires).

API : CRUD prestataires, demandes devis.

2.9 Intégrations ERP / compta (phase 2+)

Objectif : Connecter le SaaS aux systèmes existants.

Fonctionnalités :

API connecteurs (Sage, EBP, QuickBooks).

Synchronisation devis/factures.

Tech : Microservices connecteurs spécifiques.

2.10 Mobile offline-first (phase 3)

Objectif : Utilisable sur chantier sans réseau.

Fonctionnalités :

Mode offline (stockage local).

Sync automatique quand réseau dispo.

Upload photo/audio directement lié projet.

Tech : Angular + Capacitor/Ionic.

👉 En résumé :

MVP = Auth + Doc Storage + Communication + Projet/Tâches.

Phase 2 = Budget, OCR, Intégrations.

Phase différenciation = IA vision, Marketplace, Mobile offline.