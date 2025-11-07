# Summary

`cloud_helper` is a CLI tool for managing and monitoring PostgreSQL instances across multiple cloud providers (AWS RDS, Google Cloud SQL, Azure, OVHcloud).

# Features

## Generate

The `generate` subcommand generates configuration files (`postgresql.conf`, `pg_hba.conf`, `.pgpass`, `audit`, `all`) from templates.

Usage:
```sh
cloud_helper generate [flags]
```

Options:
- `--file`: File to generate (`postgresql.conf`, `pg_hba.conf`, `.pgpass`, `audit`, `all`)
- `--template`: Template to use for generation
- `--template-name`: Template name to use for generation (default "audit.md")

More info with `generate --help`

## Pgbadger

The `pgbadger` subcommand generates pgbadger reports from downloaded logs.

Usage:
```sh
cloud_helper pgbadger [flags]
```

Options:
- `--input`: Input files or directory (REQUIRED)
- `--log-line-prefix`: Log line prefix (default `"%m [%p]: [%l-1] xact=%x,user=%u,db=%d,client=%h, app=%a"`)
- `--output`: Output directory (default `"./pgbadgers"`)

More info with `pgbadger --help`

## Quellog

The `quellog` subcommand provides advanced PostgreSQL log analysis.

Usage:
```sh
cloud_helper quellog [flags]
```

Options:
- `--input`: Input files or directory (REQUIRED)
- `--output`: Output directory (empty = stdout)
- `--summary`: Show summary only
- `--checkpoints`: Show checkpoints
- `--connections`: Show connections
- `--sql-performance`: Show SQL performance
- `--json`: Export as JSON
- `--md`: Export as Markdown

More info with `quellog --help`

## GCP

Interact with Google Cloud SQL PostgreSQL.

Global options:
- `--project-id`: Google Cloud project ID

### list

List all instances in the specified project.

Usage:
```sh
cloud_helper gcp list [flags]
```

### psql

Connect to a Cloud SQL instance using IAM authentication.

Usage:
```sh
cloud_helper gcp psql [flags]
```

Options:
- `--dbname`: Database name
- `--host`: Instance connection name (format: `project:region:instance`)
- `--iam`: Use IAM authentication
- `--username`: Username

## RDS

Interact with AWS RDS PostgreSQL.

Global options:
- `--db-instance-identifier`: RDS instance identifier
- `--list`: List all RDS instances
- `--profile`: AWS profile to use (default "default")

### download

Download RDS logs or metrics.

Usage:
```sh
cloud_helper rds download [flags]
```

Options:
- `--directory`: Destination directory (default `"./"`)
- `--end`: End date (default `"2025/02/27 17:56:00"`)
- `--start`: Start date (default `"2025/02/26 17:56:00"`)
- `--type`: File type to download (`logs`, `metrics`) (default `"logs"`)

### psql

Execute subcommands or queries.

Usage:
```sh
cloud_helper rds psql [flags]
```

#### show

Display PostgreSQL settings.

Usage:
```sh
cloud_helper rds psql show [settings...] [flags]
```

Options:
- `-f, --force`: Don't ask for confirmation to retrieve all settings
- `-F, --format`: Output format (`default`, `csv`, `json`) (default `"default"`)

## Azure

Interact with Azure Database.

Global options:
- `--account-name`: Azure account to use (default `"default"`)

### download

Download files from an Azure container within a time range.

Usage:
```sh
cloud_helper azure download [flags]
```

Options:
- `--begin-time`: Start time to filter files (required, format: `YYYY-MM-DD HH:MM:SS`)
- `--container-name`: Azure Blob container name (required)
- `--end-time`: End time to filter files (required, format: `YYYY-MM-DD HH:MM:SS`)

## OVH

Interact with OVHcloud Database PostgreSQL.

Global options:
- `--service-name`: OVHcloud Public Cloud Service Name (Project ID)
- `--cluster-id`: OVHcloud Database Cluster ID
- `--endpoint`: OVHcloud API endpoint (`ovh-eu`, `ovh-ca`, `ovh-us`)

### list

List all PostgreSQL databases.

Usage:
```sh
cloud_helper ovh list [flags]
```

### psql

Connect to an OVHcloud Database instance.

Usage:
```sh
cloud_helper ovh psql [flags]
```

# Examples

## Generate pgbadger report

```bash
cloud_helper pgbadger --input=<input-file-or-directory> --log-line-prefix="<log-line-prefix>" --output=<output-directory>
```

## RDS

### List instances

Currently, you need at least one instance identifier to retrieve others:

```bash
cloud_helper rds --profile=<my-profile> --list
```

### Download logs

```bash
cloud_helper rds --profile=<my-profile> --db-instance-identifier=<my-id> download --type=logs --start="2025/02/10 08:00:00" --end="2025/02/10 09:00:00"
```

## Azure

### Download files from Azure container

```bash
cloud_helper azure --account-name=<account-name> download --container-name=<container-name> --begin-time="2025-02-10 08:00:00" --end-time="2025-02-10 09:00:00"
```

## GCP

### List instances in specified project

```bash
cloud_helper gcp --project-id=<project-id> list
```

### Connect to Cloud SQL instance

```bash
cloud_helper gcp --project-id=<project-id> psql --host=<instance-connection-name> --dbname=<database-name> --username=<username>
```

## OVH

### List databases

```bash
cloud_helper ovh --service-name=<service-name> list
```

### Connect to database

```bash
cloud_helper ovh --service-name=<service-name> --cluster-id=<cluster-id> psql
```

# Installation

Download the package from one of the [releases](https://github.com/robinportigliatti/cloud_helper/releases).

## From binaries

Download the binary for your platform from the releases page.

## From source

```bash
git clone https://github.com/robinportigliatti/cloud_helper.git
cd cloud_helper
go build .
```

## Via go install

```bash
go install github.com/robinportigliatti/cloud_helper@latest
```
