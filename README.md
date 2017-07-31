# sink

A directory synchronization solution between multiple clients. Still under development.


## How to build

Note that the whole project should be located inside your `$GOPATH`. There is more information at [golang.org](https://golang.org/doc/install#tarball).

- Client side

  The client can be built by typing `go build` in the root directory, which will generate the `./sink` executable file.

- Server side

  The server code is in the `./server` directory. Simply typing `go build` command in `./server` folder to build a server binary. The `server` executable will be generated.


## How to use

- Go to `./server` folder and run `./server` command to start the server
- In the root folder of this repo, run `./sink testDir <your unique device id>` to start the client.
