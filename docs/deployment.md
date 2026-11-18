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
cp config.example.yaml /opt/rtsprec/config.yaml
chown -R rtsprec:rtsprec /opt/rtsprec /var/cache/rtsprec
chmod 755 /opt/rtsprec/rtsprec
chmod 600 /opt/rtsprec/config.yaml
```

### 4. Configure Service
```bash
cp deployments/rtsprec.service /etc/systemd/system/
systemctl daemon-reload
systemctl enable rtsprec
systemctl start rtsprec
```

### 5. Verify Service
```bash
systemctl status rtsprec
curl http://localhost:8080/api/cameras
```

## Rollback
```bash
cp /opt/rtsprec.old/rtsprec /opt/rtsprec/
systemctl restart rtsprec
```

## Configuration Checklist
- [ ] config.yaml created with S3 credentials
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
