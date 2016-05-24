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

and watch the logs.

The mainfest includes the other necessary environment variables that will control the scaling of the application based on memory usage.  These are:

```
APPLICATION_NAME:  the name of the application that will need to be scaled by this nozzle

MEMORY_THRESHOLD:  The upper limit of the memory usage of the application that will trigger scaling behaviour.

TIME_OVER_THRESHOLD:  The amount of time that an application must be over the MEMORY_THRESHOLD limit before scaling can be triggered.

TIME_BETWEEN_SCALES:  The amount of time that this nozzle will wait between scaling activities

```

The basic algorithm for scaling is based on the average memory usage for all instances of the given application.  In order to acheive this, the nozzle keeps a basic map of <instance id>/<memory struct> where the <memory struct> is a combination of memory usage as returned in the container metric and the last time the metric was collected.

This 'last time the metric was collected' is important when the number of application instances scales down.  Rather than keep track of which instances have been removed in a scale-down situation, and realizing that the nth instance may not be removed in an n-instance collection (it may be the zeroth, or second, but is not guaranteed to be the nth) that `last time` portion of the value of the map is used in the memory algorithm.

The algorithm for determining whether or not to scale up based on memory usage is as follows:

1.  Every time a ContainerMetric event is received by the nozzle,
2.  Annotate that event with application and container data
3.  Using this annotated event, determine if the event is for the application in question by examining the [app_name] event
4.  If the event is for the application in question, add a new value to the MemoryMap collection:
	the key for the map entry is the instance id of the container
	the value is the struct:
		memory is the [memory_bytes] field of the event
		last time updated is <now>
5.  After the event has been added to/updated in the MemoryMap, calculate the average memory usage of all running instances:
	an instance is only considered a running instance if its last time updated is within the last ten minutes.  Applications with a last updated time that is more than ten minutes old is considered a crashed or scaled-down app and is no longer active
6.  If the average memory usage exceeds the MEMORY_THRESHOLD (all values in MB),
	check the time that the MEMORY_THRESHOLD was first exceeded.  If that time is `1`, then the MEMORY_THRESHOLD has not yet been exceeded, so set the ThresholdFirstExceeded time to current time.
	if this ThresholdFirstExeceed time is not 1, then compare it to the current time.  If that delta is greater than TIME_OVER_THRESHOLD then we have been above that memory threshold for TIME_OVER_THRESHOLD seconds and a scaling event might be appropriate.
7.  If the memory has been over the MEMORY_THRESHOLD threshold for more than TIME_OVER_THRESHOLD seconds, compare the LastScaledTime to the current time.  If that delta is greater than TIME_BETWEEN_SCALES, then invoke a scaling event.

The checks for the TIME_OVER_THRESHOLD are to insure against scaling when there are spikes in the memory profile.  The check for the TIME_BETWEEN_SCALES is to allow a new application to come up and handle requests before making another request to scale.

Currently this application only scales in an upward direction: a similar pattern could be implemented for a lower bound.

### Usage Notes

Applications that intend to use this nozzle need to be deployed to a diego cell.  Container Metrics that come from a DEA cell will not be accurate.

Further, when deploying a Java application against the standard Java Buldpack, the heap is pre-allocated (-Xms and -Xmx are the same).  This means that accurate memory usage statistics are not possible with container metrics unless forks to this buildpack for memory-specific purposes are used. 
