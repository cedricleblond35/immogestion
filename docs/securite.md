# Sécurité
OIDC/OAuth2 via Keycloak (ou Auth0) : PKCE, refresh token rotation, device flow mobile.
RBAC par organisation + ABAC (attributs : rôle, appartenance au projet).
API Gateway : JWT vérifiés, scopes/claims, rate limiting, WAF.
mTLS (SPIFFE/SPIRE via Istio) entre services, network policies.
Chiffrement : TLS 1.3, AES-256 au repos (S3 SSE), enveloppe via HashiCorp Vault pour secrets.
GDPR : droit à l’oubli, retenue de logs limitée, data residency (buckets par région).
Audit : audit logs immuables (Loki) + horodatage signé (RFC3161 si nécessaire).