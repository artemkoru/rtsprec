package capture

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"time"

	"github.com/bluenviron/gortsplib/v5"
	"github.com/bluenviron/gortsplib/v5/pkg/base"
	"github.com/bluenviron/gortsplib/v5/pkg/format"
	"github.com/pion/rtp"
)

type FrameData struct {
	Timestamp time.Time
	Data      []byte
}

type RTSPReader struct {
	cameraID  string
	rtspURL   string
	username  string
	password  string
	frameChan chan FrameData
	stopChan  chan struct{}
	logger    *slog.Logger

	// Retry logic
	maxRetries      int
	retryDelay      time.Duration
	currentRetrying int
	retryBackoff    time.Duration
}

func NewRTSPReader(cameraID, rtspURL, username, password string, frameBufferSize int) *RTSPReader {
	return &RTSPReader{
		cameraID:     cameraID,
		rtspURL:      rtspURL,
		username:     username,
		password:     password,
		frameChan:    make(chan FrameData, frameBufferSize),
		stopChan:     make(chan struct{}),
		logger:       slog.Default().With("camera_id", cameraID),
		maxRetries:   10,
		retryDelay:   2 * time.Second,
		retryBackoff: 2,
	}
}

func (r *RTSPReader) Start(ctx context.Context) {
	go r.run(ctx)
}

func (r *RTSPReader) Stop() {
	close(r.stopChan)
}

func (r *RTSPReader) FrameChannel() <-chan FrameData {
	return r.frameChan
}

func (r *RTSPReader) run(ctx context.Context) {
	defer close(r.frameChan)
	r.currentRetrying = 0

	for {
		select {
		case <-ctx.Done():
			return
		case <-r.stopChan:
			return
		default:
		}

		if err := r.connect(ctx); err != nil {
			r.handleError(ctx, err)
			continue
		}

		// Connection successful, reset retry counter
		r.currentRetrying = 0
		r.retryDelay = 2 * time.Second
	}
}

func (r *RTSPReader) connect(ctx context.Context) error {
	u, err := base.ParseURL(r.rtspURL)
	if err != nil {
		return fmt.Errorf("invalid rtsp url: %w", err)
	}

	// Credentials are carried in the URL user info in gortsplib v5
	if r.username != "" {
		(*url.URL)(u).User = url.UserPassword(r.username, r.password)
	}

	r.logger.Info("connecting to RTSP stream", "url", r.rtspURL)

	c := &gortsplib.Client{
		Scheme:      u.Scheme,
		Host:        u.Host,
		ReadTimeout: 10 * time.Second,
	}
	if err := c.Start(); err != nil {
		return fmt.Errorf("failed to start client: %w", err)
	}
	defer c.Close()

	desc, _, err := c.Describe(u)
	if err != nil {
		return fmt.Errorf("failed to describe stream: %w", err)
	}

	r.logger.Info("connected to RTSP stream")

	// Find a supported video format: H264 first, then H265
	var h264Format *format.H264
	var forma format.Format
	medi := desc.FindFormat(&h264Format)
	if medi != nil {
		forma = h264Format
	} else {
		var h265Format *format.H265
		medi = desc.FindFormat(&h265Format)
		if medi == nil {
			return fmt.Errorf("no supported video track found")
		}
		forma = h265Format
	}

	if _, err := c.Setup(desc.BaseURL, medi, 0, 0); err != nil {
		return fmt.Errorf("failed to setup video track: %w", err)
	}

	c.OnPacketRTP(medi, forma, func(pkt *rtp.Packet) {
		r.handleFrame(pkt)
	})

	if _, err := c.Play(nil); err != nil {
		return fmt.Errorf("failed to play stream: %w", err)
	}

	r.logger.Info("RTSP stream playing")

	// Block until the client fails, the context ends, or Stop is called
	waitErr := make(chan error, 1)
	go func() { waitErr <- c.Wait() }()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-r.stopChan:
		return fmt.Errorf("reader stopped")
	case err := <-waitErr:
		return fmt.Errorf("rtsp connection closed: %w", err)
	}
}

func (r *RTSPReader) handleFrame(pkt *rtp.Packet) {
	// Convert RTP packet to frame data
	frameData := FrameData{
		Timestamp: time.Now(),
		Data:      pkt.Payload,
	}

	select {
	case r.frameChan <- frameData:
	case <-r.stopChan:
	default:
		// Frame buffer full, skip frame
		r.logger.Debug("frame buffer full, skipping frame")
	}
}

func (r *RTSPReader) handleError(ctx context.Context, err error) {
	if r.currentRetrying >= r.maxRetries {
		r.logger.Error("max retries exceeded", "error", err)
		return
	}

	r.currentRetrying++
	waitTime := r.retryDelay

	r.logger.Warn("RTSP connection lost, retrying",
		"error", err,
		"attempt", r.currentRetrying,
		"max_attempts", r.maxRetries,
		"wait_seconds", waitTime.Seconds(),
	)

	select {
	case <-ctx.Done():
		return
	case <-r.stopChan:
		return
	case <-time.After(waitTime):
	}

	// Increase delay exponentially
	r.retryDelay = time.Duration(float64(r.retryDelay) * 1.5)
	if r.retryDelay > 5*time.Minute {
		r.retryDelay = 5 * time.Minute
	}
}
