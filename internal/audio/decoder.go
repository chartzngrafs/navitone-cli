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
		f.buffer = make([]byte, 0, len(frame.Subframes[0].Samples)*f.channels*2)
		f.bufPos = 0

		// Interleave channels properly
		for i := 0; i < len(frame.Subframes[0].Samples); i++ {
			for ch := 0; ch < f.channels && ch < len(frame.Subframes); ch++ {
				sample := int16(frame.Subframes[ch].Samples[i])
				// Convert to little endian bytes
				f.buffer = append(f.buffer, byte(sample&0xFF), byte(sample>>8))
			}
			// If mono, duplicate to stereo
			if f.channels == 1 {
				sample := int16(frame.Subframes[0].Samples[i])
				f.buffer = append(f.buffer, byte(sample&0xFF), byte(sample>>8))
			}
		}
	}

	return n, nil
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

		// Need to read more data - read in chunks that match our channel layout
		samplesPerChannel := 1024
		samples := make([]float32, samplesPerChannel*o.channels)
		samplesRead, err := o.reader.Read(samples)
		if err != nil {
			if err == io.EOF && n > 0 {
				return n, nil
			}
			return n, err
		}

		// Convert float32 samples to 16-bit PCM bytes
		o.buffer = make([]byte, 0, samplesRead*2) // 2 bytes per sample
		o.bufPos = 0

		for i := 0; i < samplesRead; i++ {
			// Convert float32 (-1.0 to 1.0) to int16
			floatSample := samples[i] * 32767
			// Clamp to prevent overflow
			if floatSample > 32767 {
				floatSample = 32767
			} else if floatSample < -32768 {
				floatSample = -32768
			}
			sample := int16(floatSample)
			// Convert to little endian bytes
			o.buffer = append(o.buffer, byte(sample&0xFF), byte(sample>>8))
		}

		// If mono, duplicate to stereo
		if o.channels == 1 {
			originalLen := len(o.buffer)
			for i := 0; i < originalLen; i += 2 {
				o.buffer = append(o.buffer, o.buffer[i], o.buffer[i+1])
			}
		}
	}

	return n, nil
}

// WAVDecoder handles WAV format
type WAVDecoder struct {
	sampleRate int
	channels   int
	duration   int64
}

func (d *WAVDecoder) Decode(r io.Reader) (io.Reader, error) {
	// For WAV files, we need to parse the header and skip to the data
	// For simplicity, let's assume WAV files are already in the correct format
	// and just return the reader after skipping the header

	// Note: This is a simplified implementation
	// A full implementation would parse the WAV header properly
	d.sampleRate = 44100 // Assume standard rate
	d.channels = 2       // Assume stereo
	d.duration = 0       // Unknown

	return r, nil
}

func (d *WAVDecoder) SampleRate() int { return d.sampleRate }
func (d *WAVDecoder) Channels() int   { return d.channels }
func (d *WAVDecoder) Duration() int64 { return d.duration }
