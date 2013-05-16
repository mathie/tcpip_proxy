# TCP/IP Proxy in Go

**TL;DR: I took some existing code and refactored it to the way I'd structure it.
However, this is the first chunk of Go I've ever touched. Please submit pull
requests to turn it into idiomatic Go!**

For some reason, I was reading
[The Beauty of Concurrency in Go](http://pragprog.com/magazines/2012-06/the-beauty-of-concurrency-in-go)
yesterday morning and decided to spend 20 minutes typing out the code to see
how it felt in reality. I've been meaning to try Go for a while now, but every
time I aim to give it a shot, I wind up picking too big a challenge, where I
should be focusing the effort on learning the language. So, a nice toy example.

## Development Environment

Of course, I can't really get started without tooling up a bit. First of all,
get Go installed on my (Mac) laptop with homebrew:

```shell
brew install go
```

which has, between me playing yesterday and documenting it today, helpfully
been updated to Go 1.1. So I'm going to have to re-figure what I've done to
make it work with 1.1! Which makes documenting where I started and how I got to
here a little tricky, so let's focus on how to get the environment up and
running.

The `go` tool relies on a couple of environment variables to help it figure out
where to find things and where to put things:

* `GOROOT` which is the root of the Go installation. If you're using homebrew,
  you can safely add `export GOROOT="$(brew --prefix go)"` to your bash/zsh
  profile and you're good to go.

* `GOPATH` which points to your local Go workspace. This is much like a
  workspace in Eclipse, as I understand it, in that it's one place to keep all
  your source code and their dependencies. It may well be sensible to have a
  workspace per project or per 'role', but for now I'm just dumping everything
  into a single workspace, so I've added
  `export GOPATH="${HOME}/Development/Go"` to my `~/.zshenv` for now.

So, there we have it, a working Go environment. Wait, there's just one extra
thing: editor support. I drink the vim koolaid, and there's a vim plugin
distributed with Go (if you've got the homebrew version installed, you'll find
it at `/usr/local/opt/go/misc/vim`). Add that plugin to vim in your preferred
manner. I added a couple of extra files:

* [`compiler/go.vim`](https://github.com/mathie/.vim/blob/master/compiler/go.vim)
  which sets the correct `makeprg` and `errorformat` when editing go programs;
  and

* [`ftplugin/go/compiler.vim`](https://github.com/mathie/.vim/blob/master/ftplugin/go/compiler.vim)
  which tells it to use the go compiler as defined above when editing go files.

Now when I run `:make` et al in vim while editing a go file, it does some
approximation of the right thing.

## A note about workspaces

I only discovered Go's workspaces this morning when I tried to build my code
against the newly installed Go 1.1 and it didn't work. The setup and layout of
workspaces is covered in detail in
[How to write Go code](http://golang.org/doc/code.html), so here's the short
version.

There are three directories inside your workspace:

* `bin` where executable commands are placed after they've been built;

* `pkg` where static libraries of your dependencies are placed after they've
  been built; and

* `src` which contains the source code to your application and its dependent
  libraries (in other words, where all the action happens).

I haven't played around with dependencies much yet, but it sounds like, while
Go will happily download and install dependencies on your behalf, you don't
have much control over the upstream versions of these dependencies. So I'm
going to revise my previous statement about workspace-per-whatever and say you
should have a workspace per project. That workspace should be version
controlled, and the import of dependencies managed as you normally would
(submodules, subtree merge, whatever).

Anyway. If you're at the root of your workspace, you can pull in my attempt at
the TCP proxy with the following command:

```shell
go get github.com/mathie/tcpip_proxy
```

which will download it, and place it in
`${GOPATH}/src/github.com/mathie/tcpip_proxy`. You can then generate the binary
with:

```shell
go install github.com/mathie/tcpip_proxy
```

This will compile the source (and any dependencies if there were any) build the
binary and dump it in the `bin/` directory. You could run it as:

```shell
bin/tcpip_proxy
```

and it'll tell you how to get it running, but that's not terribly interesting.
It does roughly the same as the article at the start says.

If you're actively hacking on this particular module, Go's OK with that, too.
Inside your workspace, you can cd into the package and start editing from
there:

```shell
cd github.com/mathie/tcpip_proxy
```

This time when you want to build the project, you can just do:

```shell
go build
```

and it'll dump the resulting executable in your current directory (so you'll
want to gitignore that...). I believe that it will still resolve other
dependencies from your workspace and the global go root.

## What I actually did

So after all that. This was meant to be a 20 minute exercise, typing/copying
the code from the article to get a feel for it, running it and tweaking
slightly. What it turned into was a day long excursion into the world of Go,
attempting to refactor the program into a more sensible (to me) structure,
while sticking with Go's idioms.

Most of what I did was to break the code up into smaller functions, because
that's how I think. But I also tried to divide it into clumps of data and the
operations performed on that data (which sounds an awful lot like objects).
Here's the four objects I extracted:

* [`Logger`](logger.go) which encapsulates the goroutine which takes log
  messages and dumps them out to the appropriate log file.

* [`Channel`](channel.go) which encapsulates a unidirectional channel between
  two sockets, logging and forwarding packets, in another goroutine.

* [`Connection`](connection.go) which combines two channels - one in each
  direction - plus an overall logger for general connection information.

* [`Proxy`](proxy.go) which listens for new connections, then kicks off a new
  connection goroutine to handle each of them.

There's also the main program itself in [tcpip_proxy.go](tcpip_proxy.go) which
parses command line arguments, then kicks off a `Proxy` to make it all work.
It's just wiring.

Initially, I split off all these objects into separate packages, naming the
package after the single class inside, and following the convention for
exported names. (The convention is that names beginning with an uppercase
letter are exported from the package; those beginning with a lower case letter
aren't.) However, after switching to the workspace setup in Go 1.1, I've moved
them all back to a single package. I'm still a little unsure about what 'size'
a package should be, how granular things should be, what I should be exporting
from packages and suchlike. Something I'm sure will start to gel as I write
larger code bases.

So, yes, most of what I learned was how to clump together related bits of data,
and how to define behaviours on that data. In terms of the data, you define it
by creating a type which is just a new label for any built in type. So if the
'data' that you're operating on can be represented as a single string, you
could do:

```go
type Hostname string
```

However, in my cases, I was wanting to clump together a few bits of data, so my
type would typically be a label for a struct:

```go
type Proxy struct {
  target           string
  localPort        string
  connectionNumber int
}
```

Idiomatically, your package will have a constructor method to build a new one
of these things (in this case, it's trivial):

```go
func NewProxy(targetHost, targetPort, localPort string) *Proxy {
  target := net.JoinHostPort(targetHost, targetPort)

  return &Proxy{
    target:           target,
    localPort:        localPort,
    connectionNumber: connectionNumber,
  }
}
```

(I just discovered that if you're splitting a "composite literal" like that
over several lines - say because you're writing documentation and want to keep
the line lengths short - then the final line must have a trailing `,` too. I
also discovered that error messages from the Go compiler are generally very
helpful.)

The [Effective Go](http://golang.org/doc/effective_go.html) documentation also
says that if the constructor is constructing something that's obvious from the
package name, just call it `New`.

So now we've got a clump of data and a means to build it. How to we define
behaviours for it? It took me about 3 reads of Effective Go to spot it, but
this is how you define these methods:

```go
func (proxy Proxy) Run() {
  // Do stuff.
}
```

I suppose I missed it because that looks a lot like defining return types in
other languages. It's not, it's defining the type that the method operates on
and giving it a name to access it inside the method. So, inside the method, the
fields of the proxy struct are available as (e.g.) `proxy.target`, etc.

Calling the method on the data is as you'd expect:

```go
proxy := NewProxy("localhost", "4000", "5000")
proxy.Run()
```

Straightforward enough. So that's data and their operations. Effective Go pays
a lot of attention to interfaces, which seem like a related topic, but I
couldn't see any way to apply them here, so I, well, haven't.

## Packages

As I said above, I had split out all these objects into separate packages when
I was working with Go 1.0.3 yesterday, but have coalesced it back into a single
package now. As far as I can tell, a package is one of two things:

* A library which other code depends upon which, when installed will produce a
  static library in `${GOPATH}/pkg`. In this case, start out each of your files
  in the package with the 'short' package name (conventionally, the name of the
  repo it sits in). So, if I was distributing this project as a library, I'd be
  sticking `package tcpip_proxy` at the head of every file.

* And if it's not a library, it's a program which installs a binary into
  `${GOPATH}/bin`. In this case, start out each of the files in the package
  with `package main`. This makes it a program. I haven't tried, but it seems
  reasonable that you can have multiple programs (packages whose name is
  `package main`) in a single workspace which are distinct instead by their
  full import path.

The one thing I find odd about this: it doesn't seem possible to distribute a
single package which can be both a library and a binary. Say, for example, I
considered this project to be primarily a library, but I included a trivial
binary to demonstrate its use. I think I'd have to distribute the trivial
binary as a separate package. Something to investigate further another time.

## Testing

To round off my afternoon, I'm having a shot of the built in testing library
for Go, having noticed that the `go` tool has built in support for compiling
and running tests. It's pretty straightforward: normally files with the
`_test.go` suffix aren't compiled and linked into your build. However, if you
run:

```shell
go test github.com/mathie/tcpip_proxy
```

from the root of your workspace, or just:

```shell
go test
```

from the package itself, it builds a different binary (`tcpip_proxy.test`,
which you'll probably want to gitignore, too) and runs it. This invokes the
test runner which runs all your tests and reports on the results, as you'd
expect.

The testing framework feels a little unusual to my xUnit (OK, latterly rspec)
eyes. There are three sorts of tests you can run: regular tests, benchmarks and
examples. The sort of test is reflected from the test method signature.

Examples are fairly straightforward. These methods are of the form
`func ExampleFooBar()`. You execute some code and provide a comment at the end
of the method to describe the expected output on stdout. Useful for checking
the output of command line programs. For example:

```go
func ExampleOutput() {
  fmt.Println("Hello, world")

  // Output:
  // Hello, world
}
```

First thing I jumped at to test was that my program was correctly generating
some help text to describe the command line arguments. This highlighted two
interesting problems:

* It only appears to allow you to provide the output of `stdout` and
  `flag.PrintDefaults()` (correctly) prints to `stderr`.

* The test framework adds a bunch of extra flags to the running program anyway.

So I couldn't think up another useful example in this app...

Benchmarks are really neat and I spent a little while playing around with how
performance changes by splitting a trivial example over a number of goroutines.
Benchmark methods have the signature `func BenchmarkFooBar(b *testing.B)`.
Having failed to come up with a better example, I was inspired by the
documentation to benchmark `fmt.Sprintf`:

```go
func DoSprintf(times int) {
  for j := 0; j < times; j++ {
    fmt.Sprintf("This benchmark will be run %d times.\n", times)
  }
}
```

and a wee wrapper to run it as a goroutine, signalling along a channel that it's done:

```go
func GoDoSprintf(times int, signal chan bool) {
  go func(signal chan bool) {
    DoSprintf(times)

    signal <- true
  }(signal)
}
```

and finally a wrapper for that to 'split' the work amongst a number of
goroutines and wait for them all to signal that they've finished before
returning:

```go
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
```

Straightforward enough. (Incidentally, I nicked the pattern for the goroutines
to signal completion from the original program. I hope it's an idiomatic Go
pattern where it's needed; I think it's quite neat.) Now, a simple benchmark to
get a baseline without using goroutines:

```go
func BenchmarkSprintf(b *testing.B) {
  DoSprintf(b.N)
}
```

and a benchmark running with various numbers of goroutines:

```go
func BenchmarkGoroutineSprintf2(b *testing.B) {
  GoroutineSprintf(2, b.N)
}
```

I have permutations for this for 2, 4, 8 and 16 goroutines. It's pretty neat.
It runs all your benchmarks, then reports on how long each operation takes
(presumably total time / number of iterations). A few observations, though:

* It only runs the benchmarks if you specify the `-bench` flag. So to get it to
  run all your benchmarks, run `go test -bench .`.

* Even if you do specify the `-bench` flag, it will not run your benchmarks if
  you have any failing tests or examples, which it runs first. This makes
  sense, I suppose, but tripped me up because my example file had examples of
  failing tests, too!

* At first glance, goroutines didn't seem to have much performance benefit at
  all. Then I noticed the `-cpu` flag which specifies a list of `GOMAXPROCS`
  (the number of simultaneously running goroutines) values for test runs. It
  says this defaults to the 'current value of GOMAXPROCS' which in my case
  appears to be 1. It accepts a list, so specifying `1,2,4,8` means it will run
  each benchmark (and test/example too) with `GOMAXPROC` set to each of those
  values. I started to see an improvement when it was `> 1`. :)

* Not as much improvement as I'd have liked, though. It took a while to strike
  me as to why. It turns out that Go attempts to run each benchmark for
  approximately a fixed period of time (1 second by default, controlled by the
  `-benchtime` argument). This means it will repeatedly call your benchmark
  method, refining `b.N` each time so that the total runtime is about 1 second.
  (You can see this by adding `fmt.Printf("Running DoSprintf(%d).\n", times)`
  into the `DoSprintf` method.)

  I can see why this is useful, but I would like to have fixed `b.N` here so
  the benchmark was being called once, to see if the overhead of setting up the
  goroutines and channels was affecting the overall performance.

So that's the benchmarks. How about regular, 'proper' tests? Well, they're
characterised by the method signature `func TestFooBar(t *testing.T)` and
they're the ones that are a bit 'different' to me. There doesn't seem to be the
usual xUnit pattern of having assertions. Instead, a test is considered to have
passed if it runs to completion without being skipped or failed. So, here's a
passing test:

```go
func TestPassingTest(t *testing.T) {
  result := 1 + 1
  fmt.Sprintf("1 + 1 = %d\n", result)
}
```

No assertions, nor a `should` in sight, just plain old code. That feels a bit
odd. You can run tests with `go test` which will give you some terse output
about whether it passes or fails and a total runtime:

```shell
> go test
PASS
ok      github.com/mathie/tcpip_proxy   0.018s
```

or you can pass in the `-v` flag to get verbose output:

```shell
> go test -v
=== RUN TestPassingTest
--- PASS: TestPassingTest (0.00 seconds)
PASS
ok      github.com/mathie/tcpip_proxy   0.017s
```

So, how do we mark a test that should be skipped (don't run the (remainder of)
the test and note that it was skipped in the verbose output)? Like this:

```go
func TestSkippedTest(t *testing.T) {
  fmt.Println("This will be printed.")
  t.Skip("Skipping this test for now.")
  fmt.Println("This will not be printed.")
}
```

which verbosely outputs something like:

```go
=== RUN TestSkippedTest
This will be printed.
--- SKIP: TestSkippedTest (0.00 seconds)
        channel_test.go:15: Skipping this test for now.
PASS
ok      github.com/mathie/tcpip_proxy   0.017s
```

As you can see, skipped tests are considered to be passing (if you don't ask
for the verbose output, you'll get no notification that there are any skipped
tests at all). So, how about failing tests?

```go
func TestImmediatelyFailingTest(t *testing.T) {
  fmt.Println("This will be printed.")
  t.Fatal("This test will fail now and not run to completion")
  fmt.Println("This will not be printed.")
}
```

The test will fail and will immediately terminate the test method (I wonder if
it uses `panic` and `recover` under the covers? `defer` still works.) This
seems like the natural behaviour, but there is an alternative:

```go
func TestFailingTest(t *testing.T) {
  fmt.Println("This will be printed.")
  t.Error("This test will ultimately fail, but will continue to completion")
  fmt.Println("This will also be printed.")
}
```

which marks the test as failed, but still continues running it. I suppose that
would be useful if there's some state you need to restore after the test
anyway, but that seems like a prime use of `defer`... Either way, the verbose
test output looks like:

```shell
=== RUN TestFailingTest
This will be printed.
This will also be printed.
--- FAIL: TestFailingTest (0.00 seconds)
        channel_test.go:21: This test will ultimately fail, but will continue to completion
=== RUN TestImmediatelyFailingTest
This will be printed.
--- FAIL: TestImmediatelyFailingTest (0.00 seconds)
        channel_test.go:28: This test will fail now and not run to completion
FAIL
exit status 1
FAIL    github.com/mathie/tcpip_proxy   0.018s
```

Fair enough. I'll endeavour to continue mucking around with the test framework
but so far it's feeling a little "low level", like it's the tools that you'd
use to write a test framework rather than the framework you'd use. I suppose
everybody's testing needs are different (cf the sheer number of testing
frameworks in the Ruby community). I think I'd be happy with something along
the lines of the following:

```go
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
```

I wonder how easy it would be to tidy that up so `setU()` and `tearDown()` were
automatically called if defined, and I could avoid having to pass
`t *testing.T` into each `assert`...

## Formatting Code

`go fmt` automatically formats your code in the proscribed Go style. And there
is a single proscribed Go style, so you can be reasonably sure that any random
code you pick up that's written in Go is within a command of looking the same.
It also helps that apparently we largely agree on said style anyway. After two
days of hacking with the code, the only things it picked me up on where:

* Tabs, not spaces. OK, I can live with that one. My position on the tabs vs
  spaces debate is "not both".

* Imports should be in alphabetical order. Good call again, I was mostly being
  lazy. It's good to note then that the ordering plain isn't important to the
  compiler, which beats `#include` ordering games...

* No spacing out string concatenation. So when I write out
  `"[" + timestamp() + "]"`, it takes out the spaces. Less enthusiastic about
  this one (I like my horizontal space) but OK.

* Vertically aligning types in struct definitions, composite literals, etc.
  Yeah, I normally do that too, I just forgot that one time, OK?

* Not vertically aligning successive assignments. You liked me vertically
  aligning other things, I wonder if this is an oversight?

## Things I like about go

I love working with goroutines and channels for passing messages between them.
It feels like a really natural way to think about software development. And, of
course, it can allow the program to scale up onto multiple cores on your
computer, run goroutines concurrently and get things done faster. I'm sure
there are still plenty of ways to trip myself up, but I managed to write a
trivial program that I observed running with 13 separate threads and not once
have to think about the complexities of concurrent programming. That's got to
be a win.

I liked that the compiler makes me keep my imports in check, so that I can
clearly see dependencies. The number of times I've looked at the myriad of
`require 'foo'` lines in a ruby file that's been around for a few years and
wondered if they're all necessary. Or, worse, with Rails' autoloader, not even
knowing what a file's dependencies really are! This was particularly awesome as
I split bits out into separate files.

Multiple return values from a method. In particular, this comes into its own
for signalling errors. the typical idiom is to do something along the lines of:

```go
bytesRead, err := channel.Read(buffer)
if err != nil {
  panic(fmt.Sprintf("Channel read failed: %v", err))
}

// Carry on
```

This way we don't have to think of 'special' values of the return value
(idiomatically -1 in C) to indicate errors, and then pass the actual error
status out of band. I also like the pace of "call a method, check for errors,
call a method, check for errors". I always liked that style in C; apart from
anything else, it's clear to see when errors are, and aren't, being checked
without jumping out of the current context.

`defer` is neat. It schedules a method to run at the end of the current scope,
no matter how the current scope is exited. So far, most of what I've used it
for is to remember to close open files when I'm done with them - the same as
  I'd do with blocks in Ruby. So, in Ruby:

```ruby
def cracker
  File.open('/etc/passwd') do |f|
    # IN UR PASSWD FILE, CRACKIN UR PASSWDS
  end
end
```

which automatically closes the file at the end of the block. The equivalent in
Go:

```go
func cracker() {
  f, err := os.Open("/etc/passwd")
  // error checking elided...
  defer f.Close()

  // IN UR PASSWD FILE, CRACKIN UR PASSWDS
}
```

The Go version can be more flexible, because it allows the caller, rather than
the callee, to define the behaviour that happens at the end of the scope. And
it avoids an extra level of indentation, which pleases me.

Despite its unusual (to me) behaviour, I'm getting on rather well with the test
suite. We're definitely going to spend some more time together. Although it
might be time spent ... enhancing it the way I'd like to use it. :)

## Conclusion

I've run out of things to say. I've quite enjoyed this wee exercise.
Refactoring existing code has been an excellent way to learn a bit more about
Go - after all, by definition refactoring is not about introducing new
behaviour so I wasn't having to think about the problem domain. I could just
focus on finding out about bits of Go and use them to morph the program in some
way.

I'd be really interested in feedback. This is the first chunk of Go I've
written. It was all written while staring at Effective Go, but I'm sure it's
not yet idiomatic Go. (I've seen code from experienced developers new to Ruby
writing idiomatically in their preferred language while using Ruby's keywords.
I have no doubt that this code will smell of Ruby being written in Go syntax!)

Pull requests to turn it into idiomatic Go would be much appreciated.
