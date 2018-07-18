// Copyright 2018 Google Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package events

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"runtime/debug"
	"time"

	"google.golang.org/api/pubsub/v1"
	"google.golang.org/api/storage/v1"

	"../nodego"
)

// JSTime is a wrapper for time.Time to decode the time from Javascript.
type JSTime struct {
	time.Time
}

// UnmarshalJSON parses a JSON string to time.Time.
func (t *JSTime) UnmarshalJSON(b []byte) (err error) {
	if string(b) == `"null"` {
		t.Time = time.Time{}
		return
	}

	t.Time, err = time.Parse(`"2006-01-02T15:04:05.000Z"`, string(b))
	return
}

// EventContext holds the data associated with the event that triggered the
// execution of the function along with metadata of the event itself.
type EventContext struct {
	EventID   string `json:"eventId"`
	Timestamp JSTime `json:"timestamp"`
	EventType string `json:"eventType"`
	Resource  string `json:"resource"`
}

// Event is the basic data structure passed to functions by non-HTTP
// triggers.
type Event struct {
	Context EventContext
	Data    json.RawMessage
}

// UnmarshalJSON parses a JSON string to time.Time.
func (e *Event) UnmarshalJSON(b []byte) error {
	raws := map[string]json.RawMessage{}
	if err := json.Unmarshal(b, &raws); err != nil {
		return err
	}

	rawContext, ok := raws["context"]
	if !ok {
		rawContext = b
	}

	if err := json.Unmarshal(rawContext, &e.Context); err != nil {
		return err
	}

	e.Data = raws["data"]
	return nil
}

// PubSubMessage is a wrapper for pubsub.PubsubMessage.
type PubSubMessage struct {
	pubsub.PubsubMessage

	Data []byte
}

// PubSubMessage unmarshals the event data as a pub sub message.
func (e *Event) PubSubMessage() (*PubSubMessage, error) {
	var msg pubsub.PubsubMessage
	if err := json.Unmarshal(e.Data, &msg); err != nil {
		return nil, err
	}

	decoded, err := base64.StdEncoding.DecodeString(msg.Data)
	if err != nil {
		return nil, err
	}

	return &PubSubMessage{
		PubsubMessage: msg,
		Data:          decoded,
	}, nil
}

// StorageObject is a wrapper for storage.Object.
type StorageObject struct {
	storage.Object

	// TODO consider adding pre-parsed time fields.
}

// StorageObject unmarshals the event data as a storage event.
func (e *Event) StorageObject() (*StorageObject, error) {
	var obj storage.Object
	if err := json.Unmarshal(e.Data, &obj); err != nil {
		return nil, err
	}

	return &StorageObject{
		Object: obj,
	}, nil
}

// Handler returns http.Handler that parses the body for a function event.
func Handler(handler func(*Event) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO potentially extract information from the request path.
		//
		// PubSub and Bucket triggers have the following request path
		// structures respectively:
		//
		// /execute/_ah/push-handlers/pubsub/projects/{PROJECT_NAME}/topics/{TOPIC_NAME}
		// /execute/_ah/push-handlers/pubsub/projects/{ARBITRARY_VALUE}/topics/{ARBITRARY_VALUE}
		//
		// It seems that, for the time being, bucket triggers are actually just
		// pubsub triggers internally.

		// TODO flush logs before sending response, as in worker.js

		defer func() {
			if r := recover(); r != nil {
				w.WriteHeader(http.StatusInternalServerError)
				nodego.ErrorLogger.Printf("%s:\n\n%s\n", r, debug.Stack())
			}
		}()

		defer r.Body.Close()

		var event Event
		if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
			nodego.ErrorLogger.Print("Failed to decode event: ", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err := handler(&event); err != nil {
			nodego.ErrorLogger.Print(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
