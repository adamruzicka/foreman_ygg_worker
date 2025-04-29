package main

import (
	"context"
	"errors"
	"strings"
	"syscall"
	"testing"

	"github.com/subpop/go-log"
)

func TestDispatch(t *testing.T) {
	jobStorage := newJobStorage()

	tests := []struct {
		description string
		input       Message
		want        string
	}{
		{
			want: "Received unknown event 'test'\n",
			input: Message{
				Metadata: map[string]string{
					"event": "test",
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			var writer strings.Builder

			log.SetOutput(&writer)
			log.SetFlags(0)

			sc := serverContext{jobStorage: &jobStorage}

			dispatch(context.Background(), test.input, &sc)

			got := writer.String()
			if got != test.want {
				t.Errorf("|%+v| != |%+v|", got, test.want)
			}
		})
	}
}

func TestCancel(t *testing.T) {
	jobStorage := newJobStorage()
	jobStorage.Set("123-abc", 12345)

	tests := []struct {
		description string
		input       Message
		killError   error
		want        string
	}{
		{
			want: "No job uuid found in job's metadata, aborting.\n",
		},
		{
			want: "Cannot cancel unknown job 456-qwe\n",
			input: Message{
				Metadata: map[string]string{
					"job_uuid": "456-qwe",
				},
			},
		},
		{
			want: "Cancelling job 123-abc, sending SIGTERM to process 12345\n",
			input: Message{
				Metadata: map[string]string{
					"job_uuid": "123-abc",
				},
			},
		},
		{
			want: "Cancelling job 123-abc, sending SIGTERM to process 12345\n" +
				"Failed to send SIGTERM to process 12345: kill: (12345) - No such process\n",
			input: Message{
				Metadata: map[string]string{
					"job_uuid": "123-abc",
				},
			},
			killError: errors.New("kill: (12345) - No such process"),
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			var writer strings.Builder

			log.SetOutput(&writer)
			log.SetFlags(0)
			log.SetLevel(log.LevelInfo)

			syscallKill = func(pid int, sig syscall.Signal) error { return test.killError }
			defer func() { syscallKill = syscall.Kill }()

			sc := serverContext{jobStorage: &jobStorage}

			cancel(context.Background(), test.input, &sc)

			got := writer.String()
			if got != test.want {
				t.Errorf("|%+v| != |%+v|", got, test.want)
			}
		})
	}
}
