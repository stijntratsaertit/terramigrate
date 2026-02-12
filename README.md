<div align="center">

# terramigrate

**Declarative database schema management for PostgreSQL**

Define your desired database schema in YAML, and terramigrate computes the diff against your live database, generates migration files, and applies them transactionally.

</div>

## Installation

```bash
go install stijntratsaertit/terramigrate@latest
```

Or build from source:

```bash
git clone <repo-url>
cd terramigrate
make build
```

## Configuration

Create a `.env` file (or set environment variables):

```env
DATABASE_HOST=localhost
DATABASE_PORT=5432
DATABASE_NAME=mydb
DATABASE_USER=postgres
DATABASE_PASSWORD=secret
```

## Usage

### 1. Define your desired schema

Create a `db.yaml` file describing your desired database state:

```yaml
namespaces:
  - name: public
    tables:
      - name: users
        columns:
          - name: id
            type: INTEGER
            nullable: false
            default: "nextval('users_id_seq')"
            primary_key: true
          - name: email
            type: CHARACTER VARYING
            max_length: 255
            nullable: false
            default: "''"
        constraints:
          - name: users_pkey
            type: PRIMARY KEY
            targets: [id]
        indices:
          - name: idx_users_email
            unique: true
            algorithm: btree
            columns: [email]
    sequences:
      - name: users_id_seq
        type: bigint
```

### 2. Plan a migration

```bash
terramigrate plan --file db.yaml --description "initial schema"
```

This compares your `db.yaml` against the live database and generates migration files in `./migrations/`:

```
migrations/
  20260212_143000_initial_schema/
    up.sql      # Forward migration SQL
    down.sql    # Auto-generated rollback SQL
    plan.yaml   # Metadata (version, checksum, etc.)
```

### 3. Apply pending migrations

```bash
terramigrate apply
```

Shows a summary and prompts for confirmation. For CI pipelines:

```bash
terramigrate apply --auto-approve
```

### 4. Check migration status

```bash
terramigrate status
```

### 5. Rollback

```bash
terramigrate rollback --steps 1
```

### Other commands

```bash
terramigrate show       # Print the current live database state
terramigrate export     # Export current DB state to a YAML file
```

## Commands

| Command    | Description                                         |
| ---------- | --------------------------------------------------- |
| `plan`     | Diff desired state and generate migration files     |
| `apply`    | Execute pending migrations                          |
| `rollback` | Reverse the last N applied migrations               |
| `status`   | Show applied/pending migration status               |
| `show`     | Print the current live database state               |
| `export`   | Export the current database state to a YAML file    |

## Global Flags

| Flag              | Default      | Description                    |
| ----------------- | ------------ | ------------------------------ |
| `-a, --adapter`   | `postgres`   | Database adapter to use        |

## Contributing

1. Fork the repository
2. Create a feature branch
3. Run tests: `make test`
4. Submit a pull request

### License

This project is licensed under the [Apache License, Version 2.0](LICENSE).
