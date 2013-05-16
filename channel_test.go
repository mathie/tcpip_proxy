package tcpip_proxy

import (
  "fmt"
  "testing"
)

func setUp() {
  fmt.Println("Shared test setup for this module.")
}

func tearDown() {
  fmt.Println("Shared test teardown for this module.")
}

func assert(t *testing.T, value bool, message string) {
  if !value {
    t.Fatalf("Assertion failed: %v should have been true but was not: %v", value, message)
  }
}

func TestPassingTestWithAssert(t *testing.T) {
  setUp()
  defer tearDown()

  assert(t, true, "True is not true!")
}

func TestFailingTestWithAssert(t *testing.T) {
  setUp()
  defer tearDown()

  assert(t, false, "False is not true!")
}

func TestPassingTest(t *testing.T) {
  result := 1 + 1
  fmt.Sprintf("1 + 1 = %d\n", result)
}

func TestSkippedTest(t *testing.T) {
  fmt.Println("This will be printed.")
  t.Skip("Skipping this test for now.")
  fmt.Println("This will not be printed.")
}

func TestFailingTest(t *testing.T) {
  fmt.Println("This will be printed.")
  t.Error("This test will ultimately fail, but will continue to completion")
  fmt.Println("This will also be printed.")
}

func TestImmediatelyFailingTest(t *testing.T) {
  fmt.Println("This will be printed.")
  defer fmt.Println("I wonder if this will be printed")
  t.Fatal("This test will fail now and not run to completion")
  fmt.Println("This will not be printed.")
}

func DoSprintf(times int) {
  fmt.Printf("Running DoSprintf(%d).\n", times)

  for j := 0; j < times; j++ {
    fmt.Sprintf("This benchmark will still be run %d times.\n", times)
  }
}

func GoDoSprintf(times int, signal chan bool) {
  go func(signal chan bool) {
    DoSprintf(times)

    signal <- true
  }(signal)
}

func GoroutineSprintf(goRoutines, times int) {
  signal := make(chan bool)
  share := times / goRoutines

  for i := 0; i < goRoutines; i++ {
    GoDoSprintf(share, signal)
  }

  for i := 0; i < goRoutines; i++ {
    <-signal
  }
}

// Benchmarks only run if the test suite passes *and* you run
// `go test -bench=.` to switch them on.
func BenchmarkSprintf(b *testing.B) {
  DoSprintf(b.N)
}

func BenchmarkGoroutineSprintf2(b *testing.B) {
  GoroutineSprintf(2, b.N)
}

func BenchmarkGoroutineSprintf4(b *testing.B) {
  GoroutineSprintf(4, b.N)
}

func BenchmarkGoroutineSprintf8(b *testing.B) {
  GoroutineSprintf(8, b.N)
}

func BenchmarkGoroutineSprintf16(b *testing.B) {
  GoroutineSprintf(16, b.N)
}

func ExampleOutput() {
  fmt.Println("Hello, world")

  // Output:
  // Hello, world
}
