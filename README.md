Voici la version entièrement traduite en français, en respectant ta consigne :

---

# Résumé

`cloud_helper` est un outil vous permettant de récupérer des informations sur vos instances AWS RDS.

# Fonctionnalités

## Génération

La sous-commande `generate` permet de générer un fichier `postgresql.conf`, `pg_hba.conf`, `.pgpass`, `audit`, `all` en fonction d'un `template`.

Utilisation :
```sh
cloud_helper generate [flags]
```

Options :
- `--file` : Fichier à générer (`postgresql.conf`, `pg_hba.conf`, `.pgpass`, `audit`, `all`)
- `--template` : Template à utiliser pour la génération
- `--template-name` : Nom du template à utiliser pour la génération (par défaut "audit.md")

Plus d'infos avec `generate --help`

## Pgbadger

La sous-commande `pgbadger` permet :

1. De générer un rapport `pgbadger` en fonction de ce qui a été téléchargé par la sous-commande `download`.

Utilisation :
```sh
cloud_helper pgbadger [flags]
```

Options :
- `--input` : Fichiers ou répertoire d'entrée (OBLIGATOIRE)
- `--log-line-prefix` : Préfixe de ligne de log (par défaut `"%m [%p]: [%l-1] xact=%x,user=%u,db=%d,client=%h, app=%a"`)
- `--output` : Répertoire de sortie (par défaut `"./pgbadgers"`)

Plus d'infos avec `pgbadger --help`

## GCP

Interagir avec Google Cloud Project SQL Cloud PostgreSQL.

Options globales :
- `--project-id` : Identifiant du projet Google Cloud

### list

Lister toutes les instances dans le projet spécifié.

Utilisation :
```sh
cloud_helper gcp list [flags]
```

### psql

Se connecter à une instance Cloud SQL en utilisant IAM.

Utilisation :
```sh
cloud_helper gcp psql [flags]
```

Options :
- `--dbname` : Nom de la base de données
- `--host` : Nom de connexion de l'instance (format : `projet:region:instance`)
- `--iam` : Utiliser l'authentification IAM
- `--username` : Nom d'utilisateur

## RDS

Interagir avec AWS RDS PostgreSQL.

Options globales :
- `--db-instance-identifier` : Identifiant de l'instance RDS
- `--list` : Lister toutes les instances RDS
- `--profile` : Profil AWS à utiliser (par défaut "default")

### download

Télécharger les journaux ou les métriques RDS.

Utilisation :
```sh
cloud_helper rds download [flags]
```

Options :
- `--directory` : Répertoire de destination (par défaut `"./"`)
- `--end` : Date de fin (par défaut `"2025/02/27 17:56:00"`)
- `--start` : Date de début (par défaut `"2025/02/26 17:56:00"`)
- `--type` : Type de fichier à télécharger (`logs`, `metrics`) (par défaut `"logs"`)

### psql

Exécuter des sous-commandes ou une requête.

Utilisation :
```sh
cloud_helper rds psql [flags]
```

#### show

Afficher les paramètres PostgreSQL.

Utilisation :
```sh
cloud_helper rds psql show [settings...] [flags]
```

Options :
- `-f, --force` : Ne pas demander de confirmation pour récupérer tous les paramètres
- `-F, --format` : Format de sortie (`default`, `csv`, `json`) (par défaut `"default"`)

## Azure

Interagir avec le stockage Azure.

Options globales :
- `--account-name` : Compte Azure à utiliser (par défaut `"default"`)

### download

Télécharger des fichiers depuis un conteneur Azure sur une plage de temps définie.

Utilisation :
```sh
cloud_helper azure download [flags]
```

Options :
- `--begin-time` : Heure de début pour filtrer les fichiers (obligatoire, format : `YYYY-MM-DD HH:MM:SS`)
- `--container-name` : Nom du conteneur Azure Blob (obligatoire)
- `--end-time` : Heure de fin pour filtrer les fichiers (obligatoire, format : `YYYY-MM-DD HH:MM:SS`)

# Démo

## Génération du rapport pgbadger

```bash
cloud_helper pgbadger --input=<fichier-ou-répertoire-d'entrée> --log-line-prefix="<préfixe-ligne-log>" --output=<répertoire-de-sortie>
```

## RDS

### Récupération de la liste des instances

Pour le moment, il faut au moins un identifiant d'instance pour récupérer les autres… :

```bash
cloud_helper rds --profile=<mon-profil> --list
```

### Récupération des journaux

```bash
cloud_helper --profile=<mon-profil> --db-instance-identifier=<mon-id> rds download --type=logs --start="2025/02/10 08:00:00" --end="2025/02/10 09:00:00"
```

## Azure

### Téléchargement des fichiers d'un conteneur Azure

```bash
cloud_helper --account-name=<nom-compte> azure download --container-name=<nom-conteneur> --begin-time="2025-02-10 08:00:00" --end-time="2025-02-10 09:00:00"
```

## GCP

### Liste des instances dans le projet spécifié

```bash
cloud_helper gcp --project-id=<id-projet> list
```

### Connexion à une instance Cloud SQL

```bash
cloud_helper gcp --project-id=<id-projet> psql --host=<nom-connexion-instance> --dbname=<nom-base-données> --username=<nom-utilisateur>
```

# Installation

Récupérer le paquet dans l'une des releases.