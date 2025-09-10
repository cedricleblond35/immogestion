# Entrer ds le conteneur

sudo docker exec -it immogestion_postgres_dev bash

# se connecter à Postgres

psql -h localhost -U immobilier_user -d immobilier_prod

\dn
\dt auth.\*
\d auth.users

select \* from auth.users;
