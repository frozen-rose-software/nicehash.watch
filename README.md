# NiceHash.Watch

NiceHash.Watch is a simple watcher application that polls the NiceHash API every 5 minutes to see if your number of workers has changed.  When it does notice a change, it uses Twilio to send a text message.  This can be useful to monitor when your miner crashes or bluescreens and doesn't automatically come back up.

## Getting Started

These instructions will get NiceHash.Watch up and running for testing purposes, and then pushed to a CloudFoundry container such as Bluemix.

### Prerequisites

What you need:

```
golang
godep
GNU make

Twilio Account
NiceHash Wallet Address
```

### Building

First, copy `config.sample.yml` to `config.yml` and configure it accordingly.  Then:

```
make
```


### Running

Simply run:

```
./nicehash.watch
```

### Shutting Down

Navigate to your service (e.g. [http://localhost:8080](http://localhost:8080)) and enter your `shutdownSecret` that was configured in `config.yml`.

> **Warning:** Shutting down NiceHash.Watch will *only* shutdown the monitor.  Your miner will continue to run.

### Deploying

Assuming you are deploying to a CloudFoundry service (such as Bluemix), copy `manifest.sample.yml` to `manifest.yml`.  Then:

```
cf push
```

## Authors

* **Craig St. Jean** - [craigstjean](https://github.com/craigstjean)

## License

This project is licensed under the MIT License - see the [LICENSE.md](LICENSE.md) file for details
