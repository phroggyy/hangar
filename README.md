# Hangar

Avoid the tens of port bindings you easily end up with if you run several docker-based web projects on your local machine.
We've all been there, when port 80 is taken, so you just add the mapping `81:80`, and then `82:80`, and 83, 84, 85...

Hangar is a convenient wrapper around the Caddy webserver, used as a reverse proxy for all your other webservers. It will
automatically issue a self-signed certificate, and will make your project available on `{name}.test` where `{name}` is the
name of your web container.

## Prerequisites

At the moment, this does not ship with any of the dependencies. Thus, to build hangar, you'll have to run `go get` for:

```
github.com/docker/docker
github.com/docker/go-connections
```

Once you have that, simply run `go build` and then `./hangar`
