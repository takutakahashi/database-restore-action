# database-restore-action
GitHub Actions for checking if database backup can be restored and restored database worked properly

## Usage

### 0. Install required command

Below commands are required:

- `mysql`
- `mysqldump`
- `aws` (if use s3 backup)

### 1. Compose configuration file

See below for more details.

```
database:
  type: mysql
  name: test
  user: root
  password: root
  host: 127.0.0.1
  port: 33060
  disable_redo_log: false
check:
- query: "select * from test"
  operator: exists
backup:
  uri: file://./sample/test.dump
  local: ./sample/test.dump
```

### database

`database` section specifies db settings backup file will be restored.
Basic environment is as a container on the actions runner.

### check

`check` section specifies test query and result verification operator.

### backup

`backup` section specifies how backup file will be fetched.
`local`, `s3`, `scp` are supported.

#### local

```local

backup:
  local: ./sample/test.dump
```

#### s3

```s3
# sample for s3://test-bucket/dir/test.dump
backup:
  s3:
    bucket: test-bucket
    key: dir/test.dump
    profile: my-project # if not defined, use default profile or env
```

#### scp

Supported private-key and password authentication and key passphrase.
Required `SSH_KEY` or `SSH_PASSWORD`, Optional `SSH_PASSPHRASE`.

```scp
backup:
  scp:
    host: 192.168.0.11
    port: 2222
    user: backupuser
```

## 2. Set actions workflow

Like below:

```
      - name: backup test
        uses: takutakahashi/database-restore-action@main
        with:
          config-path: .github/workflows/config.yaml
```

# Future works

- [ ] PostgreSQL
- [ ] Other backup sources
