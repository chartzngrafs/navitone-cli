package audio

import (
	"fmt"
	"io"
	"strings"

	"github.com/hajimehoshi/go-mp3"
	"github.com/jfreymuth/oggvorbis"
	"github.com/mewkiz/flac"
)

// Decoder interface for audio format decoders
type Decoder interface {
	Decode(r io.Reader) (io.Reader, error)
	SampleRate() int
	Channels() int
	Duration() int64 // Duration in samples, 0 if unknown
}

// DecoderFactory creates appropriate decoder based on file format
func NewDecoder(format string) (Decoder, error) {
	format = strings.ToLower(format)

	switch format {
	case "mp3":
		return &MP3Decoder{}, nil
	case "flac":
		return &FLACDecoder{}, nil
	case "ogg", "oga":
		return &OGGDecoder{}, nil
	case "wav", "wave":
		return &WAVDecoder{}, nil
	default:
		return nil, fmt.Errorf("unsupported audio format: %s", format)
	}
}

// MP3Decoder handles MP3 format
type MP3Decoder struct {
	sampleRate int
	channels   int
	duration   int64
}

func (d *MP3Decoder) Decode(r io.Reader) (io.Reader, error) {
	decoder, err := mp3.NewDecoder(r)
	if err != nil {
		return nil, fmt.Errorf("failed to create MP3 decoder: %w", err)
	}

	d.sampleRate = decoder.SampleRate()
	d.channels = 2 // go-mp3 always outputs stereo
	d.duration = decoder.Length()

	// The go-mp3 decoder already outputs the correct format:
	// - 16-bit signed little endian
	// - Stereo (2 channels)
	// - Correct sample rate
	return decoder, nil
}

func (d *MP3Decoder) SampleRate() int { return d.sampleRate }
func (d *MP3Decoder) Channels() int   { return d.channels }
func (d *MP3Decoder) Duration() int64 { return d.duration }

// FLACDecoder handles FLAC format
type FLACDecoder struct {
	sampleRate int
	channels   int
	duration   int64
}

func (d *FLACDecoder) Decode(r io.Reader) (io.Reader, error) {
	stream, err := flac.New(r)
	if err != nil {
		return nil, fmt.Errorf("failed to create FLAC decoder: %w", err)
	}

	info := stream.Info
	d.sampleRate = int(info.SampleRate)
	d.channels = int(info.NChannels)
	d.duration = int64(info.NSamples)

	// Use a simpler approach - let the FLAC library handle the decoding
	// and return a reader that converts it to the expected format
	return &FLACReader{stream: stream, channels: d.channels}, nil
}

func (d *FLACDecoder) SampleRate() int { return d.sampleRate }
func (d *FLACDecoder) Channels() int   { return d.channels }
func (d *FLACDecoder) Duration() int64 { return d.duration }

// FLACReader wraps flac.Stream to implement io.Reader with proper format conversion
type FLACReader struct {
	stream   *flac.Stream
	channels int
	buffer   []byte
	bufPos   int
}

func (f *FLACReader) Read(p []byte) (n int, err error) {
	for n < len(p) {
		// If we have buffered data, use it first
		if f.bufPos < len(f.buffer) {
			copied := copy(p[n:], f.buffer[f.bufPos:])
			f.bufPos += copied
			n += copied
			continue
		}

		// Need to read more data
		frame, err := f.stream.ParseNext()
		if err != nil {
			if err == io.EOF && n > 0 {
				return n, nil
			}
			return n, err
		}

		// Convert frame to 16-bit PCM bytes
		samplesPerFrame := len(frame.Subframes[0].Samples)
		// Always output stereo (2 channels)
		f.buffer = make([]byte, 0, samplesPerFrame*2*2) // samples * 2 channels * 2 bytes per sample
		f.bufPos = 0

		// Process each sample in the frame
		for i := 0; i < samplesPerFrame; i++ {
			// Handle stereo or mono input
			if f.channels >= 2 && len(frame.Subframes) >= 2 {
				// Stereo input - use left and right channels
				leftSample32 := frame.Subframes[0].Samples[i]
				rightSample32 := frame.Subframes[1].Samples[i]
				
				// Convert int32 to int16 with proper clamping
				leftSample16 := clampInt32ToInt16(leftSample32)
				rightSample16 := clampInt32ToInt16(rightSample32)
				
				// Write left channel (little endian)
				f.buffer = append(f.buffer, byte(leftSample16&0xFF), byte(leftSample16>>8))
				// Write right channel (little endian)
				f.buffer = append(f.buffer, byte(rightSample16&0xFF), byte(rightSample16>>8))
			} else {
				// Mono input - duplicate to stereo
				sample32 := frame.Subframes[0].Samples[i]
				sample16 := clampInt32ToInt16(sample32)
				
				// Write same sample to both left and right channels
				f.buffer = append(f.buffer, byte(sample16&0xFF), byte(sample16>>8))
				f.buffer = append(f.buffer, byte(sample16&0xFF), byte(sample16>>8))
			}
		}
	}

	return n, nil
}

// clampInt32ToInt16 properly converts int32 to int16 with clamping to prevent overflow
func clampInt32ToInt16(sample32 int32) int16 {
	if sample32 > 32767 {
		return 32767
	} else if sample32 < -32768 {
		return -32768
	}
	return int16(sample32)
}

// OGGDecoder handles OGG Vorbis format
type OGGDecoder struct {
	sampleRate int
	channels   int
	duration   int64
}

func (d *OGGDecoder) Decode(r io.Reader) (io.Reader, error) {
	decoder, err := oggvorbis.NewReader(r)
	if err != nil {
		return nil, fmt.Errorf("failed to create OGG Vorbis decoder: %w", err)
	}

	d.sampleRate = int(decoder.SampleRate())
	d.channels = int(decoder.Channels())
	d.duration = 0 // Duration not available from oggvorbis decoder

	return &OGGReader{reader: decoder, channels: d.channels}, nil
}

func (d *OGGDecoder) SampleRate() int { return d.sampleRate }
func (d *OGGDecoder) Channels() int   { return d.channels }
func (d *OGGDecoder) Duration() int64 { return d.duration }

// OGGReader wraps oggvorbis.Reader to implement io.Reader with proper format conversion
type OGGReader struct {
	reader   *oggvorbis.Reader
	channels int
	buffer   []byte
	bufPos   int
}

func (o *OGGReader) Read(p []byte) (n int, err error) {
	for n < len(p) {
		// If we have buffered data, use it first
		if o.bufPos < len(o.buffer) {
			copied := copy(p[n:], o.buffer[o.bufPos:])
			o.bufPos += copied
			n += copied
			continue
		}

		// Need to read more data - read in chunks
		samplesPerChannel := 1024
		samples := make([]float32, samplesPerChannel*o.channels)
		samplesRead, err := o.reader.Read(samples)
		if err != nil {
			if err == io.EOF && n > 0 {
				return n, nil
			}
			return n, err
		}

		if samplesRead == 0 {
			continue
		}

		// Convert float32 samples to 16-bit PCM bytes
		// Always output stereo (2 channels)
		numSamplesPerChannel := samplesRead / o.channels
		o.buffer = make([]byte, 0, numSamplesPerChannel*2*2) // samples * 2 channels * 2 bytes
		o.bufPos = 0

		for i := 0; i < numSamplesPerChannel; i++ {
			if o.channels >= 2 {
				// Stereo input - process left and right channels
				leftFloat := samples[i*o.channels]
				rightFloat := samples[i*o.channels+1]
				
				leftSample := clampFloat32ToInt16(leftFloat)
				rightSample := clampFloat32ToInt16(rightFloat)
				
				// Write left channel (little endian)
				o.buffer = append(o.buffer, byte(leftSample&0xFF), byte(leftSample>>8))
				// Write right channel (little endian)
				o.buffer = append(o.buffer, byte(rightSample&0xFF), byte(rightSample>>8))
			} else {
				// Mono input - duplicate to stereo
				monoFloat := samples[i]
				sample := clampFloat32ToInt16(monoFloat)
				
				// Write same sample to both left and right channels
				o.buffer = append(o.buffer, byte(sample&0xFF), byte(sample>>8))
				o.buffer = append(o.buffer, byte(sample&0xFF), byte(sample>>8))
			}
		}
	}

	return n, nil
}

// clampFloat32ToInt16 properly converts float32 to int16 with clamping to prevent overflow
func clampFloat32ToInt16(floatSample float32) int16 {
	// Convert float32 (-1.0 to 1.0) to int16
	scaled := floatSample * 32767.0
	// Clamp to prevent overflow
	if scaled > 32767.0 {
		return 32767
	} else if scaled < -32768.0 {
		return -32768
	}
	return int16(scaled)
}

// WAVDecoder handles WAV format
type WAVDecoder struct {
	sampleRate int
	channels   int
	duration   int64
}

func (d *WAVDecoder) Decode(r io.Reader) (io.Reader, error) {
	// For WAV files, we need to parse the header and skip to the data
	// For now, return an error since WAV parsing is complex and can cause issues
	return nil, fmt.Errorf("WAV format not properly supported yet - use MP3/FLAC/OGG instead")
	
	// Note: This is disabled to prevent static noise from malformed WAV handling
	// A full implementation would parse the WAV header properly and validate format
}

func (d *WAVDecoder) SampleRate() int { return d.sampleRate }
func (d *WAVDecoder) Channels() int   { return d.channels }
func (d *WAVDecoder) Duration() int64 { return d.duration }
