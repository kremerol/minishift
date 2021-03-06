/*
Copyright 2016 The Kubernetes Authors All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package util

import (
	"testing"
	"time"

	"github.com/pkg/errors"
)

// Returns a function that will return n errors, then return successfully forever.
func errorGenerator(n int, retryable bool) func() error {
	errorCount := 0
	return func() (err error) {
		if errorCount < n {
			errorCount += 1
			e := errors.New("Error!")
			if retryable {
				return &RetriableError{Err: e}
			} else {
				return e
			}

		}

		return nil
	}
}

func TestErrorGenerator(t *testing.T) {
	errors := 3
	f := errorGenerator(errors, false)
	for i := 0; i < errors-1; i++ {
		if err := f(); err == nil {
			t.Fatalf("Error should have been thrown at iteration %v", i)
		}
	}
	if err := f(); err == nil {
		t.Fatalf("Error should not have been thrown this call!")
	}
}

func TestRetry(t *testing.T) {
	f := errorGenerator(4, true)
	if err := Retry(5, f); err != nil {
		t.Fatalf("Error should not have been raised by retry.")
	}

	f = errorGenerator(5, true)
	if err := Retry(4, f); err == nil {
		t.Fatalf("Error should have been raised by retry.")
	}
}

func TestMultiError(t *testing.T) {
	m := MultiError{}

	m.Collect(errors.New("Error 1"))
	m.Collect(errors.New("Error 2"))

	err := m.ToError()
	expected := `Error 1
Error 2`
	if err.Error() != expected {
		t.Fatalf("%s != %s", err, expected)
	}

	m = MultiError{}
	if err := m.ToError(); err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
}

func TestVersionOrdinal(t *testing.T) {
	if VersionOrdinal("v3.4.1.10") < VersionOrdinal("v3.4.1.2") {
		t.Fatal("Expected 'false' Got 'true'")
	}
}

var durationTests = []struct {
	in   time.Duration
	want string
}{
	{10*time.Second + 555*time.Millisecond, "10.6s"},
	{10*time.Second + 555*time.Millisecond, "10.6s"},
	{10*time.Second + 500*time.Millisecond, "10.5s"},
	{10*time.Second + 499*time.Millisecond, "10.5s"},
	{9*time.Second + 401*time.Millisecond, "9.4s"},
	{9*time.Second + 456*time.Millisecond, "9.46s"},
	{9*time.Second + 445*time.Millisecond, "9.45s"},
	{1 * time.Second, "1s"},
	{859*time.Millisecond + 445*time.Microsecond, "859.4ms"},
	{859*time.Millisecond + 460*time.Microsecond, "859.5ms"},
	{859*time.Microsecond + 100*time.Nanosecond, "900µs"},
	{45 * time.Nanosecond, "45ns"},
}

func TestFriendlyDuration(t *testing.T) {
	for _, tt := range durationTests {
		got := FriendlyDuration(tt.in)
		expected, _ := time.ParseDuration(tt.want)
		if got != expected {
			t.Errorf("Expected %v but got %v", tt.in, got, expected)
		}
	}
}
