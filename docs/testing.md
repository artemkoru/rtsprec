# Testing Guide

## Integration Testing

### Full Pipeline Test
1. Start RTSP mock server
2. Configure camera with RTSP URL
3. Verify motion detection triggers
4. Verify video chunks are encoded
5. Verify chunks are uploaded to S3
6. Verify archive can be queried
7. Verify HLS stream is playable
8. Verify timeline displays events

### Real RTSP Stream Test
- Use real IP camera if available
- Monitor CPU usage (should stay <70% for 2 cameras)
- Monitor memory usage (should stay <1.5GB)
- Verify continuous recording for 24+ hours
- Verify no frame drops or stalls

## Performance Testing

### CPU Profiling
```bash
go tool pprof http://localhost:8080/debug/pprof/profile
```

### Memory Monitoring
- Watch RSS during 24-hour continuous recording
- Ensure no memory leaks in goroutines
- Verify HLS segment buffer doesn't grow unbounded

### S3 Upload Performance
- Monitor upload times for each chunk
- Verify retry logic works on network failures
- Test with different network conditions

### Concurrent Viewer Test
- Start 5+ HLS streams simultaneously
- Verify system remains responsive
- Monitor CPU impact

## Load Testing Recommendations
- 2 cameras × 30fps × 1080p = ~300Mbps network
- Expect 1.5GB/day per camera at 1080p
- Motion detection: 1 analysis per 15 frames = ~2 FPS per camera
- Target: <100MB memory, <70% CPU for 2-camera setup

## Test Checklist
- [ ] Integration test passes
- [ ] Real RTSP stream runs 24+ hours
- [ ] CPU stays <70% under typical load
- [ ] Memory stays <1.5GB
- [ ] HLS serving 5+ concurrent viewers
- [ ] Motion detection sensitivity is tunable
- [ ] Screenshots are captured on schedule
- [ ] Archive queries work within 100ms
- [ ] S3 uploads complete within 60s
- [ ] Graceful shutdown takes <5s
