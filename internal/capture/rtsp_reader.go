package capture

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/bluenviron/gortsplib/v4"
	"github.com/bluenviron/gortsplib/v4/pkg/base"
	"github.com/bluenviron/gortsplib/v4/pkg/formats"
	"github.com/bluenviron/gortsplib/v4/pkg/frames"
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
	conn := gortsplib.NewClient(gortsplib.ClientOptions{})
	defer conn.Close()

	r.logger.Info("connecting to RTSP stream", "url", r.rtspURL)

	// Set up authentication
	if r.username != "" {
		conn.SetUsername(r.username)
		conn.SetPassword(r.password)
	}

	// Connect with timeout
	connectCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	_, err := conn.Describe(connectCtx, r.rtspURL)
	if err != nil {
		return fmt.Errorf("failed to describe stream: %w", err)
	}

	r.logger.Info("connected to RTSP stream")

	// Set up frame handling
	var videoTrackIndex int
	videoTrack := conn.GetMediaByKind(base.StreamTypeVideo)
	if videoTrack == nil {
		return fmt.Errorf("no video track found")
	}

	// Get video codec info
	videoFormat := videoTrack.GetFormat(formats.H264{})
	if videoFormat == nil {
		videoFormat = videoTrack.GetFormat(formats.H265{})
		if videoFormat == nil {
			return fmt.Errorf("unsupported video codec")
		}
	}

	// Set up frame receiver
	conn.OnPacketRTP(videoTrackIndex, videoTrack, func(pkt *rtp.Packet) {
		r.handleFrame(pkt, videoTrack)
	})

	// Play the stream
	playCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	_, err = conn.Play(playCtx, nil)
	if err != nil {
		return fmt.Errorf("failed to play stream: %w", err)
	}

	r.logger.Info("RTSP stream playing")

	// Keep connection alive until context is done or stop is called
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-r.stopChan:
			return fmt.Errorf("reader stopped")
		default:
		}

		// Check if connection is still alive by reading from conn
		time.Sleep(100 * time.Millisecond)
	}
}

func (r *RTSPReader) handleFrame(pkt *rtp.Packet, track *gortsplib.MediaEntity) {
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
