# Deployment Guide

## Deployment Steps

### 1. Build the Binary
```bash
cd cmd/rtsprec
go build -o /tmp/rtsprec
```

### 2. Set Up System User
```bash
useradd -s /sbin/nologin rtsprec
```

### 3. Install Binary and Config
```bash
mkdir -p /opt/rtsprec/data /var/cache/rtsprec
cp /tmp/rtsprec /opt/rtsprec/
chown -R rtsprec:rtsprec /opt/rtsprec /var/cache/rtsprec
chmod 755 /opt/rtsprec/rtsprec
```

Put the settings from [Environment Variables](#environment-variables) into an environment file readable only by the service user, for example `/etc/rtsprec/env` with `chmod 600`.

### 4. Configure Service

Create `/etc/systemd/system/rtsprec.service` with `EnvironmentFile=/etc/rtsprec/env` and `ExecStart=/opt/rtsprec/rtsprec`, then:
```bash
systemctl daemon-reload
systemctl enable rtsprec
systemctl start rtsprec
```

### 5. Verify Service
```bash
systemctl status rtsprec
curl http://localhost:8080/api/cameras
```

## Environment Variables

The binary reads its configuration only from environment variables. The `-config` flag and YAML file are no longer used.

| Variable | Default | Description |
|---|---|---|
| `DB_PATH` | `./rtsprec.db` | SQLite database file |
| `HTTP_PORT` | `8080` | HTTP listen port |
| `LOG_LEVEL` | `info` | Log level |
| `ENVIRONMENT` | `development` | Deployment environment name |
| `ENCRYPTION_KEY_PATH` | `/opt/rtsprec/.encryption.key` | Key file for encrypted config store |
| `HLS_OUTPUT_PATH` | `/var/cache/rtsprec/hls` | Directory for HLS segments |
| `HLS_SEGMENT_DURATION_SECONDS` | `2` | HLS segment length |
| `HLS_SEGMENTS_PER_LIST` | `5` | Segments listed in each playlist |
| `AUTH_METHOD` | `basic` | `basic` or `oidc` |
| `AUTH_BASIC_USERNAME` | `admin` | Basic auth username |
| `AUTH_BASIC_PASSWORD` | `changeme` | Basic auth password (change it) |
| `AUTH_OIDC_PROVIDER` | `google` | `google` or `yandex` (when `AUTH_METHOD=oidc`) |
| `AUTH_OIDC_CLIENT_ID` | empty | OIDC client ID |
| `AUTH_OIDC_SECRET` | empty | OIDC client secret |
| `S3_BUCKET` | `rtsprec-videos` | S3 bucket for recordings |
| `S3_REGION` | `us-east-1` | S3 region |
| `S3_PATH_PREFIX` | `surveillance` | Key prefix inside the bucket |
| `S3_ACCESS_KEY_ID` | empty | S3 access key ID |
| `S3_SECRET_ACCESS_KEY` | empty | S3 secret access key |

## Rollback
```bash
cp /opt/rtsprec.old/rtsprec /opt/rtsprec/
systemctl restart rtsprec
```

## Configuration Checklist
- [ ] Environment file created with S3 credentials
- [ ] OIDC or Basic Auth configured
- [ ] Cameras added in web UI
- [ ] Motion sensitivity tuned for environment
- [ ] Screenshot schedule set (if needed)
- [ ] Service is running and auto-restarting
- [ ] Logs are being captured

## Monitoring
- Check systemd logs: `journalctl -u rtsprec -f`
- Monitor S3 uploads: `aws s3 ls s3://bucket/surveillance/`
- Check disk usage: `df -h /opt/rtsprec/data`
- Monitor HLS segments: `ls -la /var/cache/rtsprec/hls/`

## Troubleshooting

### Camera Offline
1. Check RTSP URL and credentials
2. Verify network connectivity to camera
3. Check logs: `journalctl -u rtsprec | grep camera_id`

### S3 Upload Failures
1. Verify AWS credentials
2. Check bucket permissions
3. Verify network connectivity to S3
4. Review logs for specific error codes

### High CPU Usage
1. Increase frame skip (reduce motion detection frequency)
2. Disable screenshot scheduling
3. Reduce video resolution if supported by camera

### Memory Leaks
1. Monitor with `ps aux | grep rtsprec`
2. Check goroutine count: `curl localhost:8080/debug/pprof/goroutine`
3. Restart service if memory grows unbounded
