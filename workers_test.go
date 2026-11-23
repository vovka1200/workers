package workers

import (
	"fmt"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"
)

// checkResults читает из results столько значений, сколько в want, и сравнивает их с want
// без учёта порядка: при нескольких воркерах результаты приходят в порядке завершения задач
func checkResults(t *testing.T, results ResultChan, want []any) {
	t.Helper()
	expected := make(map[string]int, len(want))
	for _, w := range want {
		expected[fmt.Sprintf("%#v", w)]++
	}
	for range want {
		key := fmt.Sprintf("%#v", <-results)
		if expected[key] == 0 {
			t.Errorf("unexpected result = %s", key)
			continue
		}
		expected[key]--
	}
	for key, n := range expected {
		if n > 0 {
			t.Errorf("missing result = %s", key)
		}
	}
}

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

	want := make([]any, 0, len(tests))
	for _, tt := range tests {
		want = append(want, tt.want)
	}
	checkResults(t, brigade.Results, want)

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

	want := make([]any, 0, len(tests))
	for _, tt := range tests {
		want = append(want, tt.want)
	}
	checkResults(t, brigade.Results, want)

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

	want := make([]any, 0, len(tests))
	for _, tt := range tests {
		want = append(want, tt.want)
	}
	checkResults(t, brigade.Results, want)

	brigade.Close()

}

func TestBrigadeCloseTwice(t *testing.T) {
	brigade := NewBrigade[int, int](2, func(id int, n int) int {
		return n * n
	})
	brigade.Start()

	brigade.Tasks <- 3
	if result := <-brigade.Results; result != 9 {
		t.Errorf("got = %v, want = 9", result)
	}

	brigade.Close()
	brigade.Close()
}
