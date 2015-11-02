# sleepwalk

Sleepwalk is a tool to schedule ElasticSearch settings using a simple template system consisting of time range and setting pairs:

```
08:00-16:00
{ "transient": { "cluster.routing.allocation.cluster_concurrent_rebalance": 3 } }
16:00-08:00
{ "transient": { "cluster.routing.allocation.cluster_concurrent_rebalance": 15 } }
```

A single template can hold any number of time and setting pairs, typically with each template representing a related configuration bundle (e.g. only allow 3 shard rebalances during the day, but 15 over night).

Templates (file formatted accordingly and ending in .conf) are picked up from the specified `-templates` directory on start. Every `-interval` seconds, each template is validated and any settings that are applicable according to the current time are applied in top-down order.

Time ranges are interpreted to span days. Setting 08:00-16:00 will span 8AM - 4PM of each day, while 16:00-08:00 will span 4PM until 8AM the next day.

Template files receive basic validation to ensure you've entered syntactically correct time ranges and valid json settings, but doesn't prevent you from setting things like an hour value of 25 or making nonsense API calls to ElasticSearch.

```
Usage of ./sleepwalk:
  -address string
    	ElasticSearch Address (default "http://localhost:9200")
  -interval int
    	Update interval in seconds (default 300)
  -templates string
    	Template path (default "./templates")
```

```
2015/10/28 10:42:31 Sleepwalk Running
2015/10/28 10:42:31 Reading template: rebalance.conf
2015/10/28 10:42:31 Pushing setting from template: rebalance.conf. Current settings: {"persistent":{},"transient":{"cluster":{"routing":{"allocation":{"cluster_concurrent_rebalance":"0"}}}}}
2015/10/28 10:42:31 New settings: {"persistent":{},"transient":{"cluster":{"routing":{"allocation":{"cluster_concurrent_rebalance":"3"}}}}}
```
