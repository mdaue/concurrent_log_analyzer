# Concurrent Log Analyzer
Meant to concurrently analyze log files with the following formatting:
```
log_format = '%(asctime)s | %(levelname)-8s | %(name)s:%(funcName)s:%(lineno)d - %(message)s'
date_format = '%Y-%m-%d %H:%M:%S.%f'
```

To run this log analysis tool:
1. `go build .`
2. `./concurrent_log_analyzer logs/*.log`

This assumes that log files reside in the logs directory, are free of ANSI coloring characters and end with the extension .log
