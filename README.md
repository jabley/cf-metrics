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

For each application, these will be emitted as:

```json
{
	"zone":"PWS",
	"space":"development",
	"app":"hello-python-web",
	"type":"metric",
	"timestamp":"2016-11-09T10:56:06.711449498Z",
	"Stats":{
		"0":{
			"Stats":{
        "disk_quota":536870912,
        "mem_quota":402653184,
        "Usage":{
          "CPU":0.006985241502010499,
          "Disk":186060800,
          "Mem":338554880
        }
      }
		},
		"1":{
			"Stats":{
        "disk_quota":536870912,
        "mem_quota":402653184,
        "Usage":{
          "CPU":0.007506122659531552,
          "Disk":186060800,
          "Mem":256016384
        }
      }
		},
		"2":{
			"Stats":{
        "disk_quota":536870912,
        "mem_quota":402653184,
        "Usage":{
          "CPU":0.004021443931726263,
          "Disk":186056704,
          "Mem":317755392
        }
      }
		},
		"3":{
			"Stats":{
        "disk_quota":536870912,
        "mem_quota":402653184,
        "Usage":{
          "CPU":0.00431447831668318,
          "Disk":186073088,
          "Mem":250630144
        }
      }
		}
	}
}
```

Since these are point-in-time values, we can happily collect and aggregate
those by running `cf-metrics` in multiple locations, and feed all of the logs
into a single location.

It will also poll every 1 minute, looking for events that have happened for each
application (such as process crash, restart) since cf-metrics began polling.

For each application, these events will be emitted as:

```json
{
  "zone":"PWS",
  "space":"development",
  "app":"hello-python-web",
  "type":"event",
  "timestamp":"2016-11-09T11:49:53.575738294Z",
  "EventInfo":{
    "type":"app.crash",
    "timestamp":"2016-09-20T16:21:29Z"
  }
}
```
