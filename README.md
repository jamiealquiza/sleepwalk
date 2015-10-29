# sleepwalk

Sleepwalk is a tool to schedule ElasticSearch settings using a simple template system consisting of time range and setting pairs:

```
08:00-16:00
{ "transient": { "cluster.routing.allocation.node_initial_primaries_recoveries": 5 } }
16:00-23:00
{ "transient": { "cluster.routing.allocation.node_initial_primaries_recoveries": 0 } }
```

A single template can hold any number of time and setting pairs, typically with each template representing a related configuration bundle (e.g. only allow 3 shard rebalances during the day, but 10 over night).

Templates (file formatted accordingly and ending in .conf) are picked up from the specified `-templates` directory on start. Every `-interval` seconds, each template is validated and any settings that are applicable according to the current time are applied in top-down order.

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
2015/10/28 10:42:31 Reading template: recoveries.conf
2015/10/28 10:42:31 Pushing setting from template: recoveries.conf. Current settings: {"persistent":{},"transient":{"cluster":{"routing":{"allocation":{"node_initial_primaries_recoveries":"5"}}}}}
2015/10/28 10:42:31 New settings: {"persistent":{},"transient":{"cluster":{"routing":{"allocation":{"node_initial_primaries_recoveries":"0"}}}}}
```