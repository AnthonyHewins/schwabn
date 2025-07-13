# schwabn

Schwab websockets over NATS. Spin up this server and it will connect to the schwab WS
and forward market data to NATS, which can be consumed by a consumer

## Deployment

```shell
prodctx="" # your production NATS context

go install github.com/nats-io/natscli/nats@latest # or, visit https://github.com/nats-io/natscli and install
nats context select $prodctx
./nats/create.sh # creates all NATS resources
make schwabn # makes binary -> ./bin/schwabn
./bin/schwabn # will run bin
```

Now, set the required env vars, outlined in `.env.tmpl`. Then, source those vars, and `./bin/schwabn` will run the binary