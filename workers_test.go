package workers

import (
	log "github.com/sirupsen/logrus"
	"reflect"
	"testing"
	"time"
)

func TestBrigadeStruct(t *testing.T) {
	type testTask struct {
		value string
	}
	type testResult struct {
		value string
	}
	testFunc := func(id int, task testTask) testResult {
		log.WithFields(log.Fields{
			"worker": id,
			"task":   task.value,
		}).Info("test")
		return testResult{
			task.value,
		}
	}

	type args struct {
		testTask
	}
	tests := []struct {
		name    string
		args    args
		want    testResult
		wantErr bool
	}{
		{
			name: "test 1",
			args: args{
				testTask{"hello world 1"},
			},
			want: testResult{value: "hello world 1"},
		},
		{
			name: "test 2",
			args: args{
				testTask{"hello world 2"},
			},
			want: testResult{value: "hello world 2"},
		},
	}

	brigade := NewBrigade[testTask, testResult](len(tests), testFunc)
	brigade.Start()

	for _, tt := range tests {
		brigade.Tasks <- tt.args.testTask
	}

	for _, tt := range tests {
		result := <-brigade.Results
		if !reflect.DeepEqual(result, tt.want) {
			t.Errorf("got = %v,\nwant = %v", result, tt.want)
		}
	}

	brigade.Close()

}

func TestBrigadeString(t *testing.T) {
	testFunc := func(id int, task string) string {
		log.WithFields(log.Fields{
			"worker": id,
			"task":   task,
		}).Info("test")
		return task
	}

	tests := []struct {
		name    string
		args    string
		want    string
		wantErr bool
	}{
		{
			name: "test 1",
			args: "hello world 1",
			want: "hello world 1",
		},
		{
			name: "test 2",
			args: "hello world 2",
			want: "hello world 2",
		},
	}

	brigade := NewBrigade[string, string](len(tests), testFunc)
	brigade.Start()

	for _, tt := range tests {
		brigade.Tasks <- tt.args
	}

	for _, tt := range tests {
		result := <-brigade.Results
		if !reflect.DeepEqual(result, tt.want) {
			t.Errorf("got = %v,\nwant = %v", result, tt.want)
		}
	}

	brigade.Close()

}

func TestBrigadeRace(t *testing.T) {
	testFunc := func(id int, task string) string {
		log.WithFields(log.Fields{
			"worker": id,
			"task":   task,
		}).Info("test")
		time.Sleep(1 * time.Second)
		return task
	}

	tests := []struct {
		name    string
		args    string
		want    string
		wantErr bool
	}{
		{
			name: "test 1",
			args: "hello world 1",
			want: "hello world 1",
		},
		{
			name: "test 2",
			args: "hello world 2",
			want: "hello world 2",
		},
		{
			name: "test 3",
			args: "hello world 3",
			want: "hello world 3",
		},
	}

	brigade := NewBrigade[string, string](2, testFunc)
	brigade.Start()

	go func() {
		for _, tt := range tests {
			brigade.Tasks <- tt.args
		}
	}()

	for _, tt := range tests {
		result := <-brigade.Results
		if !reflect.DeepEqual(result, tt.want) {
			t.Errorf("got = %v,\nwant = %v", result, tt.want)
		}
	}

	brigade.Close()

}
