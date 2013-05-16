# TCP/IP Proxy in Go

TL;DR: I took some existing code and refactored it to the way I'd structure it.
However, this is the first chunk of Go I've ever touched. Please submit pull
requests to turn it into idiomatic Go!

For some reason, I was reading
[The Beauty of Concurrency in Go](http://pragprog.com/magazines/2012-06/the-beauty-of-concurrency-in-go)
yesterday morning and decided to spend 20 minutes typing out the code to see
how it felt in reality. I've been meaning to try Go for a while now, but every
time I aim to give it a shot, I wind up picking too big a challenge, where I
should be focusing the effort on learning the language. So, a nice toy example.

## Development Environment

Of course, I can't really get started without tooling up a bit. First of all,
get Go installed on my (Mac) laptop with homebrew:

    brew install go

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

    go get github.com/mathie/tcpip_proxy

which will download it, and place it in
`${GOPATH}/src/github.com/mathie/tcpip_proxy`. You can then generate the binary
with:

    go install github.com/mathie/tcpip_proxy

This will compile the source (and any dependencies if there were any) build the
binary and dump it in the `bin/` directory. You could run it as:

    bin/tcpip_proxy

and it'll tell you how to get it running, but that's not terribly interesting.
It does roughly the same as the article at the start says.

If you're actively hacking on this particular module, Go's OK with that, too.
Inside your workspace, you can cd into the package and start editing from
there:

    cd github.com/mathie/tcpip_proxy

This time when you want to build the project, you can just do:

    go build

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

    type Hostname string

However, in my cases, I was wanting to clump together a few bits of data, so my
type would typically be a label for a struct:

    type Proxy struct {
      target string
      localPort string
      connectionNumber int
    }

Idiomatically, your package will have a constructor method to build a new one
of these things (in this case, it's trivial):

    func NewProxy(targetHost, targetPort, localPort string) *Proxy {
      target := net.JoinHostPort(targetHost, targetPort)

      return &Proxy{
        target:           target,
        localPort:        localPort,
        connectionNumber: connectionNumber,
      }
    }

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

    func (proxy Proxy) Run() {
      // Do stuff.
    }

I suppose I missed it because that looks a lot like defining return types in
other languages. It's not, it's defining the type that the method operates on
and giving it a name to access it inside the method. So, inside the method, the
fields of the proxy struct are available as (e.g.) `proxy.target`, etc.

Calling the method on the data is as you'd expect:

    proxy := NewProxy('localhost', '4000', '5000')
    proxy.Run()

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
