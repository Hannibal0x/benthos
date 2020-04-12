package query

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/Jeffail/benthos/v3/lib/message"
	"github.com/Jeffail/benthos/v3/lib/types"
	"github.com/stretchr/testify/assert"
)

func TestDeprecatedFunctionExpressions(t *testing.T) {
	hostname, _ := os.Hostname()

	type easyMsg struct {
		content string
		meta    map[string]string
	}

	tests := map[string]struct {
		input    string
		output   string
		messages []easyMsg
		index    int
		legacy   bool
	}{
		"hostname": {
			input:  `hostname`,
			output: fmt.Sprintf(`%v`, hostname),
			index:  1,
		},
		"metadata 1": {
			input:  `metadata:foo`,
			output: `bar`,
			index:  1,
			legacy: true,
			messages: []easyMsg{
				{
					meta: map[string]string{
						"foo":    "bar",
						"baz":    "qux",
						"duck,1": "quack",
					},
				},
				{},
			},
		},
		"metadata 2": {
			input:  `metadata:bar`,
			output: "",
			messages: []easyMsg{
				{
					meta: map[string]string{
						"foo":    "bar",
						"baz":    "qux",
						"duck,1": "quack",
					},
				},
			},
		},
		"metadata 3": {
			input:  `metadata`,
			output: ``,
			messages: []easyMsg{
				{
					meta: map[string]string{
						"foo":    "bar",
						"baz":    "qux",
						"duck,1": "quack",
					},
				},
			},
		},
		"metadata 4": {
			input:  `metadata:duck,1,`,
			output: "quack",
			messages: []easyMsg{
				{
					meta: map[string]string{
						"foo":    "bar",
						"baz":    "qux",
						"duck,1": "quack",
					},
				},
			},
		},
		"metadata 5": {
			input:  `metadata:foo,1`,
			output: "bar",
			index:  0,
			messages: []easyMsg{
				{},
				{
					meta: map[string]string{
						"foo":    "bar",
						"baz":    "qux",
						"duck,1": "quack",
					},
				},
			},
		},
		"metadata 6": {
			input:  `metadata:foo`,
			output: ``,
			index:  1,
			messages: []easyMsg{
				{
					meta: map[string]string{
						"foo":    "bar",
						"baz":    "qux",
						"duck,1": "quack",
					},
				},
				{},
			},
		},
		"metadata map": {
			input:  `metadata_json_object:1`,
			output: `{"baz":"qux","duck,1":"quack","foo":"bar"}`,
			index:  0,
			messages: []easyMsg{
				{},
				{
					meta: map[string]string{
						"foo":    "bar",
						"baz":    "qux",
						"duck,1": "quack",
					},
				},
			},
		},
		"metadata map 2": {
			input:  `metadata_json_object`,
			output: `{}`,
			index:  1,
			legacy: true,
			messages: []easyMsg{
				{},
				{
					meta: map[string]string{
						"foo":    "bar",
						"baz":    "qux",
						"duck,1": "quack",
					},
				},
			},
		},
		"metadata map 3": {
			input:  `metadata_json_object`,
			output: `{"baz":"qux","duck,1":"quack","foo":"bar"}`,
			index:  1,
			messages: []easyMsg{
				{},
				{
					meta: map[string]string{
						"foo":    "bar",
						"baz":    "qux",
						"duck,1": "quack",
					},
				},
			},
		},
		"json function": {
			input:  `json_field`,
			output: `{"foo":"bar"}`,
			messages: []easyMsg{
				{content: `{"foo":"bar"}`},
				{content: `not json`},
			},
		},
		"json function 2": {
			input:  `json_field:foo`,
			output: `bar`,
			index:  1,
			legacy: true,
			messages: []easyMsg{
				{content: `{"foo":"bar"}`},
			},
		},
		"json function 3": {
			input:  `json_field:foo,1`,
			output: `bar`,
			index:  0,
			messages: []easyMsg{
				{content: `not json`},
				{content: `{"foo":"bar"}`},
			},
		},
		"json function 5": {
			input:  `json_field:foo`,
			output: `null`,
			index:  1,
			messages: []easyMsg{
				{content: `{"foo":"bar"}`},
				{content: `not json`},
			},
		},
		"json function 6": {
			input:  `json_field:foo,0`,
			output: `bar`,
			index:  1,
			messages: []easyMsg{
				{content: `{"foo":"bar"}`},
				{content: `not json`},
			},
		},
		"error function": {
			input:  `error`,
			output: `test error`,
			messages: []easyMsg{
				{meta: map[string]string{
					types.FailFlagKey: "test error",
				}},
			},
		},
		"error function 2": {
			input:  `error:1`,
			output: ``,
			messages: []easyMsg{
				{meta: map[string]string{
					types.FailFlagKey: "test error",
				}},
			},
		},
		"error function 3": {
			input:  `error:1`,
			output: `test error`,
			messages: []easyMsg{
				{},
				{meta: map[string]string{
					types.FailFlagKey: "test error",
				}},
			},
		},
		"content function": {
			input:  `content`,
			output: `foobar`,
			index:  1,
			legacy: true,
			messages: []easyMsg{
				{content: `foobar`},
				{content: `barbaz`},
			},
		},
		"content function 2": {
			input:  `content:1`,
			output: `barbaz`,
			index:  1,
			messages: []easyMsg{
				{content: `foobar`},
				{content: `barbaz`},
			},
		},
		"content function 3": {
			input:  `content`,
			output: `barbaz`,
			index:  1,
			messages: []easyMsg{
				{content: `foobar`},
				{content: `barbaz`},
			},
		},
		"batch size": {
			input:  `batch_size`,
			output: `2`,
			messages: []easyMsg{
				{}, {},
			},
		},
		"batch size 2": {
			input:  `batch_size`,
			output: `1`,
			messages: []easyMsg{
				{},
			},
		},
	}

	for name, test := range tests {
		test := test
		t.Run(name, func(t *testing.T) {
			msg := message.New(nil)
			for _, m := range test.messages {
				part := message.NewPart([]byte(m.content))
				if m.meta != nil {
					for k, v := range m.meta {
						part.Metadata().Set(k, v)
					}
				}
				msg.Append(part)
			}

			e, err := tryParse(test.input)
			if !assert.NoError(t, err) {
				return
			}

			res := e.ToString(test.index, msg, test.legacy)
			res2 := string(e.ToBytes(test.index, msg, test.legacy))

			assert.Equal(t, test.output, res)
			assert.Equal(t, test.output, res2)
		})
	}
}

func TestDeprecatedCountersFunction(t *testing.T) {
	tests := [][2]string{
		{"count:foo", "1"},
		{"count:bar", "1"},
		{"count:foo", "2"},
		{"count:foo", "3"},
		{"count:bar", "2"},
		{"count:bar", "3"},
		{"count:foo", "4"},
		{"count:foo", "5"},
		{"count:bar", "4"},
		{"count:bar", "5"},
	}

	for _, test := range tests {
		e, err := tryParse(test[0])
		if !assert.NoError(t, err) {
			continue
		}
		res := e.ToString(0, message.New(nil), false)
		assert.Equal(t, test[1], res)
	}
}

func TestDeprecatedUUIDV4Function(t *testing.T) {
	results := map[string]struct{}{}

	for i := 0; i < 100; i++ {
		e, err := tryParse("uuid_v4")
		if !assert.NoError(t, err) {
			continue
		}
		res := e.ToString(0, message.New(nil), false)
		if _, exists := results[res]; exists {
			t.Errorf("Duplicate UUID generated: %v", res)
		}
		results[res] = struct{}{}
	}
}

func TestDeprecatedTimestamps(t *testing.T) {
	now := time.Now()

	e, err := tryParse("timestamp_unix_nano")
	if !assert.NoError(t, err) {
		return
	}
	tStamp := e.ToString(0, message.New(nil), false)

	nanoseconds, err := strconv.ParseInt(tStamp, 10, 64)
	if err != nil {
		t.Fatal(err)
	}
	tThen := time.Unix(0, nanoseconds)

	if tThen.Sub(now).Seconds() > 5.0 {
		t.Errorf("Timestamps too far out of sync: %v and %v", tThen, now)
	}

	now = time.Now()
	e, err = tryParse("timestamp_unix")
	if !assert.NoError(t, err) {
		return
	}
	tStamp = e.ToString(0, message.New(nil), false)

	seconds, err := strconv.ParseInt(tStamp, 10, 64)
	if err != nil {
		t.Fatal(err)
	}
	tThen = time.Unix(seconds, 0)

	if tThen.Sub(now).Seconds() > 5.0 {
		t.Errorf("Timestamps too far out of sync: %v and %v", tThen, now)
	}

	now = time.Now()
	e, err = tryParse("timestamp_unix:10")
	if !assert.NoError(t, err) {
		return
	}
	tStamp = e.ToString(0, message.New(nil), false)

	var secondsF float64
	secondsF, err = strconv.ParseFloat(tStamp, 64)
	if err != nil {
		t.Fatal(err)
	}
	tThen = time.Unix(int64(secondsF), 0)

	if tThen.Sub(now).Seconds() > 5.0 {
		t.Errorf("Timestamps too far out of sync: %v and %v", tThen, now)
	}

	now = time.Now()
	e, err = tryParse("timestamp")
	if !assert.NoError(t, err) {
		return
	}
	tStamp = e.ToString(0, message.New(nil), false)

	tThen, err = time.Parse("Mon Jan 2 15:04:05 -0700 MST 2006", tStamp)
	if err != nil {
		t.Fatal(err)
	}

	if tThen.Sub(now).Seconds() > 5.0 {
		t.Errorf("Timestamps too far out of sync: %v and %v", tThen, now)
	}

	now = time.Now()
	e, err = tryParse("timestamp_utc")
	if !assert.NoError(t, err) {
		return
	}
	tStamp = e.ToString(0, message.New(nil), false)

	tThen, err = time.Parse("Mon Jan 2 15:04:05 -0700 MST 2006", tStamp)
	if err != nil {
		t.Fatal(err)
	}

	if tThen.Sub(now).Seconds() > 5.0 {
		t.Errorf("Timestamps too far out of sync: %v and %v", tThen, now)
	}
	if !strings.Contains(tStamp, "UTC") {
		t.Errorf("Non-UTC timezone detected: %v", tStamp)
	}
}
