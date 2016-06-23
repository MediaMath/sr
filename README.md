# sr - simple CLI wrapper around confluent schema registry api

```bash
$ export SCHEMA_REGISTRY_URL=http://10.150.254.162:10078
$ sr ls
["foo"]
$ sr add bar ~/Desktop/mt_event.json
{id:"998"}
$ sr ls
["foo", "bar"]
$ sr ls bar
[1]
$ sr ls bar 1
...schema that was added and version and name...
```
