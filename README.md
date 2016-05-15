#Memory Based Autoscaler

This is a nozzle for the Cloud FOundry Firehose component.  It will ingest container events for a given applicaiton and if a memory threshold is exceeded, use the command line interface to automatically scale the application.

## Installation 
To install this application, you will need to push the app to a cloud foundry instance.  There is a `manifest.yml` included in the project, so all that is necessary is to type:

```
cf push memory-based-autoscaler --no-start
```

The `no-start` is important because we have not yet defined the environment variables that allow the application to connect to the firehose.  We will need a set of environment variables that looks like this:

```
API_ENDPOINT:  https://api.bosh-lite.com
CF_PULL_TIME: 9999s
FIREHOSE_PASSWORD:  (this is a secret)
FIREHOSE_SUBSCRIPTION_ID:  memory-based-autoscaler
FIREHOSE_USER:  (this is a secret)
SKIP_SSL_VALIDATION:  true
```

Once these environment variables have been set for the deployed app, it can be started with simply typing 

```
cf start memory-based-autoscaler
```

and watch the logs
