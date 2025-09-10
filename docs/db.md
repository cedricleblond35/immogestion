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