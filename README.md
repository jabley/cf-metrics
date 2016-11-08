Service to monitor Cloud Foundry applications.

## Building

`> make`

## Running

```
> env ZONE_PREFIXES=PWS,CH \
PWS_USERNAME=user@example.com \
PWS_PASSWORD='some-passphrase' \
PWS_API=https://api.run.pivotal.io \
PWS_NAME=pivotal \
CH_USERNAME=user@example.com \
CH_PASSWORD='some-other-passphrase' \
CH_API=https://api.lyra-836.appcloud.swisscom.com \
CH_NAME=swisscom \
cf-metrics
```

That will run cf-metrics, pointing at PWS and Swisscom public Cloud Foundry
instances. It will poll every 10 seconds, checking all of the apps that account
can see for cpu and memory metrics.

Since these are point-in-time values, we can happily collect and aggregate
those by running `cf-metrics` in multiple locations, and feed all of the logs
into a single location.
