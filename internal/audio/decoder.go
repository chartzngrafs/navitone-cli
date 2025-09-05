package audio

import (
	"encoding/binary"
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

// FLACReader wraps flac.Stream to implement io.Reader
type FLACReader struct {
	stream *flac.Stream
}

func (f *FLACReader) Read(p []byte) (n int, err error) {
	// Read frame from FLAC stream
	frame, err := f.stream.ParseNext()
	if err != nil {
		return 0, err
	}
	
	// Convert samples to bytes (16-bit little endian)
	samples := frame.Subframes[0].Samples
	if len(frame.Subframes) > 1 {
		// Stereo - interleave samples
		for i := 0; i < len(samples) && n+3 < len(p); i++ {
			// Left channel
			binary.LittleEndian.PutUint16(p[n:], uint16(samples[i]))
			n += 2
			// Right channel
			if i < len(frame.Subframes[1].Samples) {
				binary.LittleEndian.PutUint16(p[n:], uint16(frame.Subframes[1].Samples[i]))
			} else {
				binary.LittleEndian.PutUint16(p[n:], uint16(samples[i]))
			}
			n += 2
		}
	} else {
		// Mono
		for i := 0; i < len(samples) && n+1 < len(p); i++ {
			binary.LittleEndian.PutUint16(p[n:], uint16(samples[i]))
			n += 2
		}
	}
	
	return n, nil
}

// OGGReader wraps oggvorbis.Reader to implement io.Reader  
type OGGReader struct {
	reader *oggvorbis.Reader
}

func (o *OGGReader) Read(p []byte) (n int, err error) {
	// Read float32 samples
	samples := make([]float32, len(p)/4) // Assume 16-bit stereo
	samplesRead, err := o.reader.Read(samples)
	if err != nil {
		return 0, err
	}
	
	// Convert float32 samples to 16-bit PCM bytes
	for i := 0; i < samplesRead && n+1 < len(p); i++ {
		// Convert float32 (-1.0 to 1.0) to int16
		sample := int16(samples[i] * 32767)
		binary.LittleEndian.PutUint16(p[n:], uint16(sample))
		n += 2
	}
	
	return n, nil
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
	d.channels = 2 // MP3 is typically stereo, but go-mp3 always outputs stereo
	d.duration = decoder.Length()
	
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
	
	return &FLACReader{stream: stream}, nil
}

func (d *FLACDecoder) SampleRate() int { return d.sampleRate }
func (d *FLACDecoder) Channels() int   { return d.channels }
func (d *FLACDecoder) Duration() int64 { return d.duration }

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
	
	return &OGGReader{reader: decoder}, nil
}

func (d *OGGDecoder) SampleRate() int { return d.sampleRate }
func (d *OGGDecoder) Channels() int   { return d.channels }
func (d *OGGDecoder) Duration() int64 { return d.duration }

// WAVDecoder handles WAV format
type WAVDecoder struct {
	sampleRate int
	channels   int
	duration   int64
}

func (d *WAVDecoder) Decode(r io.Reader) (io.Reader, error) {
	// Read WAV header to extract format information
	var header struct {
		RIFF        [4]byte
		FileSize    uint32
		WAVE        [4]byte
		FmtChunk    [4]byte
		FmtSize     uint32
		AudioFormat uint16
		Channels    uint16
		SampleRate  uint32
		ByteRate    uint32
		BlockAlign  uint16
		BitsPerSample uint16
	}
	
	err := binary.Read(r, binary.LittleEndian, &header)
	if err != nil {
		return nil, fmt.Errorf("failed to read WAV header: %w", err)
	}
	
	// Verify it's a WAV file
	if string(header.RIFF[:]) != "RIFF" || string(header.WAVE[:]) != "WAVE" {
		return nil, fmt.Errorf("not a valid WAV file")
	}
	
	// Skip any extra format bytes
	if header.FmtSize > 16 {
		extraBytes := make([]byte, header.FmtSize-16)
		_, err := io.ReadFull(r, extraBytes)
		if err != nil {
			return nil, fmt.Errorf("failed to skip extra format bytes: %w", err)
		}
	}
	
	// Find the data chunk
	for {
		var chunkHeader struct {
			ID   [4]byte
			Size uint32
		}
		err := binary.Read(r, binary.LittleEndian, &chunkHeader)
		if err != nil {
			return nil, fmt.Errorf("failed to read chunk header: %w", err)
		}
		
		if string(chunkHeader.ID[:]) == "data" {
			// Found data chunk, return reader starting from here
			d.sampleRate = int(header.SampleRate)
			d.channels = int(header.Channels)
			d.duration = int64(chunkHeader.Size) / int64(header.Channels) / int64(header.BitsPerSample/8)
			return r, nil
		}
		
		// Skip this chunk
		_, err = io.CopyN(io.Discard, r, int64(chunkHeader.Size))
		if err != nil {
			return nil, fmt.Errorf("failed to skip chunk: %w", err)
		}
	}
}

func (d *WAVDecoder) SampleRate() int { return d.sampleRate }
func (d *WAVDecoder) Channels() int   { return d.channels }
func (d *WAVDecoder) Duration() int64 { return d.duration }