# sleepwalk
Makes ElasticSearch do things overnight.

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