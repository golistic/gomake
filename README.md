gomake - A GNU Make (Makefile) kinda thing in Go
================================================

Copyright (c) 2023, Geert JM Vanderkelen

The `gomake` package offers functionality to create a Makefile-like application
written in Go. Since running `make` depends on tools that needs to be installed
additionally, and `go` is always available, why  not just
do `go run ./cmd/make all`.  
It will take some coding in Go, but then again: a Makefile can also be utterly
complicated and hard to grasp.

This package tries to not have any dependencies other than the
standard Go library.


Disclaimer
----------

This is not a replacement of `GNU Make`, and although it takes some ideas
and naming, it will work differently.  
This is merely something that was created in the morning on a Thursday, by
somebody who didn't want to write yet again a Makefile like it was the 90s and
install yet again something extra.  
You will find yourself wring more code, but at least it will work on every damn
platform where Go runs (and Docker for the Docker targets..)


Installation
------------

```
go get -u github.com/golistic/gomake
```

Go 1.19 and greater is supported.


Quick Start
-----------

You want to build multi-architecture Docker images and push them to your
repository on GitHub, you can do following in your project:

File `cmd/make/main.go` with following content:

```go
package main

import (
	"github.com/golistic/gomake"
)

func main() {
	gomake.RegisterTargets(&gomake.TargetDockerBuildXPush)
	gomake.Make()
}
```

Use the `help` command to find out what targets are available:

```
$ go run ./cmd/make help
Available targets:
        docker-buildx
```

Get help for the `docker-buildx` command (if you just run it, you would get an
error saying flags are required):

```
$ go run ./cmd/make docker-buildx -h
```

Now you know what to provide, and you can execute:

```
$ go run ./cmd/make docker-buildx -image doggo -tag 1.0.0 -registry ghcr.io/yourOrg
```

Stock Targets
-------------

This will grow in time, but they serve also as example to create your own
targets:

| Target        | Type               | Description                                         |
|---------------|--------------------|-----------------------------------------------------|
| clean-vendor  | TargetCleanVendor  | Removes the `vendor` folder                         |
| docker-build  | TargetDockerBuild  | Builds Docker image locally                         |
| docker-buildx | TargetDockerBuildX | Uses `buildx` of Docker to create multi-arch images |
| vendor        | TargetVendor       | Runs `go mod vendor`                                |


Customizing
-----------

Giving arguments ain't fun on the command line. To elevate this burden, you
can pass on defaults tailored for your project.

Let us look again the Docker image building for multiple platforms as shown
in the [Quickstart](#Quick-Start). The idea is to create a copy of the stock
targets from `gomake`, and set some fields.

```go
package main

import (
	"github.com/golistic/gomake"

	"github.com/kelvin-green/maik/internal/app"
)

var (
	targetVendor           = gomake.TargetVendor
	targetCleanupVendor    = gomake.TargetCleanupVendor
	targetDockerBuildXPush = gomake.TargetDockerBuildXPush
)

func main() {
	targetVendor.Name = "vendor-for-docker"
	targetVendor.Flags = map[string]any{"out": "_vendor"}
	targetCleanupVendor.Flags = map[string]any{"out": "_vendor"}

	targetDockerBuildXPush.Flags = map[string]any{
		"registry": "ghcr.io/yourOrg",
		"image":    "myapp",
		"tag":      "1.0.0",
	}
	targetDockerBuildXPush.PreTargets = []*gomake.Target{&targetVendor}
	targetDockerBuildXPush.DeferredTargets = []*gomake.Target{&targetCleanupVendor}

	gomake.RegisterTargets(&gomake.TargetDockerBuildXPush)
	gomake.Make()
}
```

1. We copy the `gomake.TargetVendor` so the that `go mod vendor` can be provided
   with an alternative folder. We do this so that other Go tools are not using
   the `vendor` folder.
2. The copy of `gomake.TargetDockerBuildXPush` is provided with a mapping
   of flags which mimics the command line flags.
3. We do not register the copied Vendor-target. This makes it not available
   as a target, but we do use it a something that needs to be executed before
   `docker-buildx`. The clean-up is deferred after execution (it is always
   executed).

Run it as before, but now without the command line arguments:

```
$ go run ./cmd/make docker-buildx
```

Extending
---------

New targets can be added. They can be simple, and they can be very complicated.
Have a peak at the source code, files `stock_go.go` and `stock_docker.go`.


License
-------

Distributed under the MIT license. See `LICENSE.md` for more information.
