package audio

import (
	"fmt"
	"os"
	"os/exec"
	"sync"
	"time"
)

// Player handles audio playback using FFplay (part of FFmpeg suite)
// FFplay provides actual audio output to speakers
type Player struct {
	filename  string
	isPlaying bool
	isPaused  bool
	stopOnce  sync.Once
	volume    float64
	position  time.Duration
	duration  time.Duration
	ffplayCmd *exec.Cmd
	stopChan  chan struct{}
	mutex     sync.RWMutex
	startTime time.Time
}

// NewPlayer creates a new audio player
func NewPlayer() *Player {
	return &Player{
		volume:   1.0,
		stopChan: make(chan struct{}),
	}
}

// LoadAudio loads an audio file (or video with audio track) for playback
func (p *Player) LoadAudio(filename string) error {
	// Check if file exists
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return fmt.Errorf("audio file not found: %s", filename)
	}

	// Check if ffplay is available
	cmd := exec.Command("ffplay", "-version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ffplay not found. Please install FFmpeg (includes ffplay) for audio playback")
	}

	p.filename = filename

	// Get audio duration using ffprobe
	duration, err := p.getAudioDuration(filename)
	if err != nil {
		p.duration = 0
	} else {
		p.duration = duration
	}

	p.position = 0

	return nil
}

// getAudioDuration uses ffprobe to get the audio duration
func (p *Player) getAudioDuration(filename string) (time.Duration, error) {
	cmd := exec.Command("ffprobe",
		"-v", "quiet",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		filename,
	)

	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}

	var seconds float64
	_, err = fmt.Sscanf(string(output), "%f", &seconds)
	if err != nil {
		return 0, err
	}

	return time.Duration(seconds * float64(time.Second)), nil
}

// Play starts audio playback using FFplay
func (p *Player) Play() error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.filename == "" {
		return fmt.Errorf("no audio file loaded")
	}

	if p.isPlaying {
		return nil
	}

	// Reset stop channel
	p.stopChan = make(chan struct{})
	p.stopOnce = sync.Once{}

	// FFplay volume is 0-100, convert from our 0.0-1.0 scale
	volumeInt := int(p.volume * 100)

	// Start ffplay for audio playback (no video display)
	p.ffplayCmd = exec.Command("ffplay",
		"-nodisp",   // No video display
		"-autoexit", // Exit when playback ends
		"-loglevel", "quiet",
		"-volume", fmt.Sprintf("%d", volumeInt),
		"-i", p.filename,
	)

	// Don't inherit stdin to avoid terminal issues
	p.ffplayCmd.Stdin = nil
	p.ffplayCmd.Stdout = nil
	p.ffplayCmd.Stderr = nil

	if err := p.ffplayCmd.Start(); err != nil {
		return fmt.Errorf("failed to start ffplay: %w", err)
	}

	p.startTime = time.Now()
	p.isPlaying = true
	p.isPaused = false

	// Monitor playback in background
	go p.monitorPlayback()

	return nil
}

// monitorPlayback tracks playback position and handles completion
func (p *Player) monitorPlayback() {
	defer func() {
		p.mutex.Lock()
		p.isPlaying = false
		p.mutex.Unlock()
	}()

	// Wait for ffplay to complete
	if p.ffplayCmd != nil {
		p.ffplayCmd.Wait()
	}
}

// Pause pauses audio playback (not fully supported with ffplay, stops instead)
func (p *Player) Pause() error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if !p.isPlaying {
		return fmt.Errorf("audio is not playing")
	}

	// FFplay doesn't support true pause, so we stop and record position
	p.position = time.Since(p.startTime)
	p.isPaused = true

	return nil
}

// Resume resumes paused audio playback
func (p *Player) Resume() error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if !p.isPaused {
		return fmt.Errorf("audio is not paused")
	}

	p.isPaused = false
	return nil
}

// Stop stops audio playback
func (p *Player) Stop() error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if !p.isPlaying && p.ffplayCmd == nil {
		return nil
	}

	p.stopOnce.Do(func() {
		close(p.stopChan)
	})

	if p.ffplayCmd != nil && p.ffplayCmd.Process != nil {
		p.ffplayCmd.Process.Kill()
		p.ffplayCmd.Wait()
	}

	p.isPlaying = false
	p.isPaused = false
	p.position = 0
	p.ffplayCmd = nil

	return nil
}

// SetVolume sets the playback volume (0.0 to 1.0)
// Note: Volume change takes effect on next Play() call
func (p *Player) SetVolume(volume float64) error {
	if volume < 0.0 || volume > 1.0 {
		return fmt.Errorf("volume must be between 0.0 and 1.0")
	}

	p.mutex.Lock()
	p.volume = volume
	p.mutex.Unlock()

	return nil
}

// GetVolume returns the current volume
func (p *Player) GetVolume() float64 {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return p.volume
}

// Seek moves to a specific position in the audio
// Note: Seek is not fully supported with ffplay, position is tracked only
func (p *Player) Seek(position time.Duration) error {
	if position < 0 || (p.duration > 0 && position > p.duration) {
		return fmt.Errorf("seek position out of range")
	}

	p.mutex.Lock()
	p.position = position
	p.mutex.Unlock()

	return nil
}

// GetPosition returns the current playback position
func (p *Player) GetPosition() time.Duration {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	if p.isPlaying && !p.isPaused {
		return time.Since(p.startTime)
	}
	return p.position
}

// GetDuration returns the total audio duration
func (p *Player) GetDuration() time.Duration {
	return p.duration
}

// IsPlaying returns true if audio is currently playing
func (p *Player) IsPlaying() bool {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return p.isPlaying && !p.isPaused
}

// IsPaused returns true if audio is paused
func (p *Player) IsPaused() bool {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return p.isPaused
}

// Close cleans up the audio player
func (p *Player) Close() error {
	p.Stop()
	p.filename = ""
	return nil
}

// GetSupportedFormats returns supported audio formats
func (p *Player) GetSupportedFormats() []string {
	return []string{".mp3", ".wav", ".ogg", ".flac", ".aac", ".m4a", ".wma"}
}
