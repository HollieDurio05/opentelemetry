package entry

import "time"

type Entry struct {
	Timestamp  time.Time
	Body       interface{}
	Attributes map[string]interface{}
	Resource   map[string]interface{}
}

func New() *Entry {
	return &Entry{
		Attributes: make(map[string]interface{}),
		Resource:   make(map[string]interface{}),
	}
}
