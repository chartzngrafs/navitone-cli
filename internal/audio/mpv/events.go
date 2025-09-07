package mpv

import (
	"time"
)

// EventType represents different types of MPV events
type EventType string

const (
	// Playback events
	EventFileLoaded     EventType = "file-loaded"
	EventStartFile      EventType = "start-file"
	EventEndFile        EventType = "end-file"
	EventPlaybackRestart EventType = "playback-restart"
	
	// Property change events
	EventPropertyChange EventType = "property-change"
	
	// Time/position events
	EventSeek           EventType = "seek"
	EventPlaybackTime   EventType = "playback-time"
	
	// State change events
	EventPause          EventType = "pause"
	EventUnpause        EventType = "unpause"
	EventIdle           EventType = "idle"
	
	// Audio events
	EventAudioReconfig  EventType = "audio-reconfig"
	
	// Client events (custom)
	EventTrackStarted   EventType = "track-started"
	EventTrackFinished  EventType = "track-finished"
	EventTrackError     EventType = "track-error"
	EventPositionUpdate EventType = "position-update"
	EventStateChange    EventType = "state-change"
)

// EndFileReason represents the reason a file ended
type EndFileReason string

const (
	EndFileReasonEOF      EndFileReason = "eof"      // File reached end
	EndFileReasonStop     EndFileReason = "stop"     // Playback was stopped
	EndFileReasonQuit     EndFileReason = "quit"     // Player is quitting
	EndFileReasonError    EndFileReason = "error"    // An error occurred
	EndFileReasonRedirect EndFileReason = "redirect" // Redirected to different file
)

// PlaybackEvent represents a processed playback event
type PlaybackEvent struct {
	Type     EventType
	TrackID  string
	Position time.Duration
	Duration time.Duration
	Data     interface{}
}

// PropertyChangeEvent represents a property change event
type PropertyChangeEvent struct {
	Name string      `json:"name"`
	Data interface{} `json:"data"`
}

// EndFileEvent represents an end-file event
type EndFileEvent struct {
	Reason   EndFileReason `json:"reason"`
	Filename string        `json:"filename,omitempty"`
}

// EventProcessor handles and processes MPV events
type EventProcessor struct {
	currentTrackID string
	eventCallback  func(PlaybackEvent)
}

// NewEventProcessor creates a new event processor
func NewEventProcessor() *EventProcessor {
	return &EventProcessor{}
}

// SetEventCallback sets the callback for processed events
func (p *EventProcessor) SetEventCallback(callback func(PlaybackEvent)) {
	p.eventCallback = callback
}

// SetCurrentTrackID sets the current track ID for event processing
func (p *EventProcessor) SetCurrentTrackID(trackID string) {
	p.currentTrackID = trackID
}

// ProcessEvent processes an MPV event and emits appropriate playback events
func (p *EventProcessor) ProcessEvent(event MPVEvent) {
	if p.eventCallback == nil {
		return
	}

	switch event.Event {
	case string(EventFileLoaded):
		p.emitEvent(EventTrackStarted, nil)
		
	case string(EventStartFile):
		p.emitEvent(EventTrackStarted, nil)
		
	case string(EventEndFile):
		// Parse end file event
		var endFileData EndFileEvent
		if event.Reason != "" {
			endFileData.Reason = EndFileReason(event.Reason)
			endFileData.Filename = event.Filename
		}
		
		switch endFileData.Reason {
		case EndFileReasonEOF:
			p.emitEvent(EventTrackFinished, endFileData)
		case EndFileReasonError:
			p.emitEvent(EventTrackError, endFileData)
		case EndFileReasonStop, EndFileReasonQuit:
			// These are normal stop events, don't emit track finished
			p.emitEvent(EventStateChange, endFileData)
		default:
			p.emitEvent(EventTrackFinished, endFileData)
		}
		
	case string(EventPause):
		p.emitEvent(EventStateChange, map[string]interface{}{
			"paused": true,
		})
		
	case string(EventUnpause):
		p.emitEvent(EventStateChange, map[string]interface{}{
			"paused": false,
		})
		
	case string(EventSeek):
		p.emitEvent(EventPositionUpdate, map[string]interface{}{
			"position": event.Position,
			"seeking":  true,
		})
		
	case string(EventPropertyChange):
		p.handlePropertyChange(event)
		
	case string(EventIdle):
		p.emitEvent(EventStateChange, map[string]interface{}{
			"idle": true,
		})
		
	case string(EventPlaybackRestart):
		p.emitEvent(EventTrackStarted, nil)
		
	case string(EventAudioReconfig):
		// Audio configuration changed, might want to update UI
		p.emitEvent(EventStateChange, map[string]interface{}{
			"audio-reconfig": true,
		})
	}
}

// handlePropertyChange processes property change events
func (p *EventProcessor) handlePropertyChange(event MPVEvent) {
	// Try to extract property change data
	propertyData := map[string]interface{}{}
	
	// MPV sends property changes with different structures
	// Handle common property names we care about
	if event.Position > 0 {
		propertyData["position"] = event.Position
	}
	if event.Duration > 0 {
		propertyData["duration"] = event.Duration
	}
	if event.Pause {
		propertyData["paused"] = true
	}
	
	if len(propertyData) > 0 {
		if _, hasPosition := propertyData["position"]; hasPosition {
			p.emitEvent(EventPositionUpdate, propertyData)
		} else {
			p.emitEvent(EventStateChange, propertyData)
		}
	}
}

// emitEvent emits a processed playback event
func (p *EventProcessor) emitEvent(eventType EventType, data interface{}) {
	event := PlaybackEvent{
		Type:    eventType,
		TrackID: p.currentTrackID,
		Data:    data,
	}
	
	// Extract position and duration if available
	if dataMap, ok := data.(map[string]interface{}); ok {
		if pos, ok := dataMap["position"].(float64); ok {
			event.Position = time.Duration(pos * float64(time.Second))
		}
		if dur, ok := dataMap["duration"].(float64); ok {
			event.Duration = time.Duration(dur * float64(time.Second))
		}
	}
	
	p.eventCallback(event)
}

// GetEventName returns a human-readable event name
func GetEventName(eventType EventType) string {
	switch eventType {
	case EventTrackStarted:
		return "Track Started"
	case EventTrackFinished:
		return "Track Finished"
	case EventTrackError:
		return "Track Error"
	case EventPositionUpdate:
		return "Position Update"
	case EventStateChange:
		return "State Change"
	case EventFileLoaded:
		return "File Loaded"
	case EventStartFile:
		return "Start File"
	case EventEndFile:
		return "End File"
	case EventSeek:
		return "Seek"
	case EventPause:
		return "Pause"
	case EventUnpause:
		return "Unpause"
	case EventIdle:
		return "Idle"
	default:
		return string(eventType)
	}
}