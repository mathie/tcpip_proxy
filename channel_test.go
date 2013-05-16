package main

import (
  "testing"
  "fmt"
)

func TestPassingTest(t *testing.T) {
  result := 1 + 1
  fmt.Sprintf("1 + 1 = %d\n", result)
}

func TestSkippedTest(t *testing.T) {
  t.Skip("Skipping this test for now.")

  fmt.Printf("This should not be printed.\n")
}

func TestFailingTest(t *testing.T) {
  /* t.Error("This test will ultimately fail, but will continue to completion") */

  fmt.Printf("This should be printed.\n")
}

func TestImmediatelyFailingTest(t *testing.T) {
  /* t.Fatal("This test will fail now and not run to completion") */

  fmt.Printf("This should not be printed.\n")
}

func DoSprintf(times int) {
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

  for i := 0; i < goRoutines; i ++ {
    <- signal
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
