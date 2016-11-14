Service to monitor Cloud Foundry applications.

Running applications in Cloud Foundry is nice, but the visibility of operational
metrics leaves a little to be desired.

This application (which itself can be run as a worker application within a Cloud
Foundry provider) polls Cloud Foundry APIs to get application metrics and events
out of Cloud Foundry and put them into your collection system of choice.

## Running

```
Usage: cf-metrics

ZONE_PREFIXES – the environment variable that contains a comma-separated list of
prefixes for ENV vars that can be used for authenticating with a Cloud Foundry
provider.

-verbose
    Enable verbose logging
-whitelist string
    A comma-separated list of app names to collect data about. If none
    specified, then defaults to all apps that the account can view.
```

For example:

```sh
> env ZONE_PREFIXES=PWS,CH PWS_USERNAME=user@example.com \
PWS_PASSWORD='some-passphrase' \
PWS_API=https://api.run.pivotal.io \
PWS_NAME=pivotal \
CH_USERNAME=user@example.com \
CH_PASSWORD='some-other-passphrase' \
CH_API=https://api.lyra-836.appcloud.swisscom.com \
CH_NAME=swisscom cf-metrics
```

That will run `cf-metrics`, pointing at PWS and Swisscom public Cloud Foundry
instances. It will poll every 10 seconds, checking all of the apps those
accounts can see, for cpu and memory metrics along with application events.

For each application, these will be emitted as:

```json
{
  "zone":"PWS",
  "space":"development",
  "app":"hello-python-web",
  "type":"metric",
  "timestamp":"2016-11-09T10:56:06.711449498Z",
  "stats":{
    "0":{
      "stats":{
        "diskQuota":536870912,
        "memQuota":402653184,
        "usage":{
          "cpu":0.006985241502010499,
          "disk":186060800,
          "mem":338554880,
          "diskUsage":0.346565247,
          "memUsage":0.84081014
        }
      }
    },
    "1":{
      "stats":{
        "diskQuota":536870912,
        "memQuota":402653184,
        "usage":{
          "cpu":0.007506122659531552,
          "disk":186060800,
          "mem":256016384,
          "diskUsage":0.34656524658203,
          "memUsage":0.63582356770833
        }
      }
    },
    "2":{
      "stats":{
        "diskQuota":536870912,
        "memQuota":402653184,
        "usage":{
          "cpu":0.004021443931726263,
          "disk":186056704,
          "mem":317755392,
          "diskUsage":0.3465576171875,
          "memUsage":0.78915405273438
        }
      }
    },
    "3":{
      "stats":{
        "diskQuota":536870912,
        "memQuota":402653184,
        "usage":{
          "cpu":0.00431447831668318,
          "disk":186073088,
          "mem":250630144,
          "diskUsage":0.34658813476563,
          "memUsage":0.62244669596354
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
application (such as process crash, restart) since `cf-metrics` began polling.

For each application, these events will be emitted as:

```json
{
  "zone":"PWS",
  "space":"development",
  "app":"hello-python-web",
  "type":"event",
  "timestamp":"2016-11-09T11:49:53.575738294Z",
  "eventInfo":{
    "type":"app.crash",
    "timestamp":"2016-09-20T16:21:29Z"
  }
}
```

## Cloud Foundry manifest

If you did want to run this application in Cloud Foundry, you could do that
starting with a manifest like:

```yml
---
applications:
- name: worker-app
  buildpack: https://github.com/cloudfoundry/binary-buildpack.git
  command: chmod +x cf-metrics-linux-amd64 && ./cf-metrics-linux-amd64 -whitelist app1,app2 || ./cf-metrics-linux-amd64 -whitelist app1,app2
  no-route: true
  health-check-type: none
  memory: 64M
  env:
    ZONE_PREFIXES: PWS,CH
    PWS_USERNAME: user@example.com
    PWS_PASSWORD: some-passphrase
    PWS_API: https://api.run.pivotal.io
    PWS_NAME: pivotal
    CH_USERNAME: user@example.com
    CH_PASSWORD: some-other-passphrase
    CH_API: https://api.lyra-836.appcloud.swisscom.com
    CH_NAME: swisscom
```

And deploy that using Concourse like:

```yml
jobs:
- name: deploy-to-prod
  plan:
  - get: gh-metrics-release
    version: { tag: v0.4.5 }
    params:
      globs:
      - cf-metrics-linux-amd64*
  - get: gh-repo
  - aggregate:
    - put: deploy-app-pws-prod
      params:
        current_app_name: {{APP_NAME}}
        manifest: gh-repo/infra/cloud-foundry/manifest.yml
        path: gh-metrics-release

resources:
- name: gh-repo
  type: git
  source:
    uri: {{GIT_REPO}}
    branch: master
    private_key: {{GIT_PRIVATE_KEY}}

- name: gh-metrics-release
  type: github-release
  source:
    user: jabley
    repository: cf-metrics
    token: {{GIT_ACCESS_TOKEN}}

- name: deploy-app-pws-prod
  type: cf-resource
  source:
    api: {{PWS_API}}
    username: {{PWS_USER}}
    password: {{PWS_PASS}}
    organization: {{PWS_ORG}}
    space: {{PWS_PROD_SPACE}}
```

## Building

```sh
> make
```
