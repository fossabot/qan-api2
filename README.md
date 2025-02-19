# qan-api2

[![Build Status](https://travis-ci.org/percona/qan-api2.svg?branch=master)](https://travis-ci.org/percona/qan-api2)
[![Go Report Card](https://goreportcard.com/badge/github.com/percona/qan-api2)](https://goreportcard.com/report/github.com/percona/qan-api2)
[![pullreminders](https://pullreminders.com/badge.svg)](https://pullreminders.com?ref=badge)
[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2Fpercona%2Fqan-api2.svg?type=shield)](https://app.fossa.io/projects/git%2Bgithub.com%2Fpercona%2Fqan-api2?ref=badge_shield)

qan-api for PMM 2.x.

## Get Report

Examples:
```bash

curl -s -X POST -d '{"period_start_from": "2019-01-01T00:00:00Z", "period_start_to": "2019-01-01T10:00:00Z", "group_by": "queryid"}' http://127.0.0.1:9922/v0/qan/GetReport | jq

curl -s -X POST -d '{"period_start_from": "2019-01-01T00:00:00Z", "period_start_to": "2019-01-01T10:00:00Z", "group_by": "client_host"}' http://127.0.0.1:9922/v0/qan/GetReport | jq

curl -X POST -s -d '{"period_start_from": "2019-01-01T00:00:00Z", "period_start_to": "2019-01-01T10:00:00Z",  "labels": [{"key": "client_host", "value": ["10.11.12.4", "10.11.12.59"]}]}' http://127.0.0.1:9922/v0/qan/GetReport | jq

curl -s -X POST -d '{"period_start_from": "2019-01-01T00:00:00Z", "period_start_to": "2019-01-01T10:00:00Z", "group_by": "client_host", "offset": 10}' http://127.0.0.1:9922/v0/qan/GetReport | jq

curl -s -X POST -d '{"period_start_from": "2019-01-01T00:00:00Z", "period_start_to": "2019-01-01T10:00:00Z", "order_by": "num_queries"}' http://127.0.0.1:9922/v0/qan/GetReport | jq

```

```bash
curl -s -X POST -d '{"period_start_from": "2019-01-01T00:00:00Z", "period_start_to": "2019-01-01T10:00:00Z", "order_by": "num_queries", "columns": ["lock_time", "sort_scan"], "group_by": "server"}' http://127.0.0.1:9922/v0/qan/GetReport | jq
 ```

```bash
curl -X POST -d '{"period_start_from": "2019-01-01T00:00:00Z", "period_start_to": "2019-01-01T10:00:00Z"}'  http://127.0.0.1:9922/v0/qan/Filters/Get
```

## Get list of availible metrics.

`curl -X POST -d '{}' http://127.0.0.1:9922/v0/qan/GetMetricsNames -s | jq`

```json
{
  "data": {
    "bytes_sent": "Bytes Sent",
    "count": "Count",
    "docs_returned": "Docs Returned",
    "docs_scanned": "Docs Scanned",
    "filesort": "Filesort",
    "filesort_on_disk": "Filesort on Disk",
    "full_join": "Full Join",
    "full_scan": "Full Scan",
    "innodb_io_r_bytes": "Innodb IO R Bytes",
    "innodb_io_r_ops": "Innodb IO R Ops",
    "innodb_io_r_wait": "Innodb IO R Wait",
    "innodb_pages_distinct": "Innodb Pages Distinct",
    "innodb_queue_wait": "Innodb Queue Wait",
    "innodb_rec_lock_wait": "Innodb Rec Lock Wait",
    "latancy": "Latancy",
    "load": "Load",
    "lock_time": "Lock Time",
    "merge_passes": "Merge Passes",
    "no_good_index_used": "No Good Index Used",
    "no_index_used": "No Index Used",
    "qc_hit": "Query Cache Hit",
    "query_length": "Query Length",
    "query_time": "Query Time",
    "response_length": "Response Length",
    "rows_affected": "Rows Affected",
    "rows_examined": "Rows Examined",
    "rows_read": "Rows Read",
    "rows_sent": "Rows Sent",
    "select_full_range_join": "Select Full Range Join",
    "select_range": "Select Range",
    "select_range_check": "Select Range Check",
    "sort_range": "Sort Range",
    "sort_rows": "Sort Rows",
    "sort_scan": "Sort Scan",
    "tmp_disk_tables": "Tmp Disk Tables",
    "tmp_table": "Tmp Table",
    "tmp_table_on_disk": "Tmp Table on Disk",
    "tmp_table_sizes": "Tmp Table Sizes",
    "tmp_tables": "Tmp Tables"
  }
}
```

## Get Query Exemples

`curl 'http://localhost:9922/v0/qan/ObjectDetails/GetQueryExample' -XPOST -d '{"filter_by":"1D410B4BE5060972","group_by":"queryid","limit":5,"period_start_from":"2018-12-31T22:00:00+00:00","period_start_to":"2019-01-01T06:00:00+00:00"}' -s | jq`

```json
{
  "query_examples": [
    {
      "example": "Ping",
      "example_format": "EXAMPLE",
      "example_type": "RANDOM"
    },
    {
      "example": "Ping",
      "example_format": "EXAMPLE",
      "example_type": "RANDOM"
    },
    {
      "example": "Ping",
      "example_format": "EXAMPLE",
      "example_type": "RANDOM"
    },
    {
      "example": "Ping",
      "example_format": "EXAMPLE",
      "example_type": "RANDOM"
    },
    {
      "example": "Ping",
      "example_format": "EXAMPLE",
      "example_type": "RANDOM"
    }
  ]
}
```

## Get metrics

`curl -X POST -s -d '{"period_start_from": "2019-01-01T00:00:00Z", "period_start_to": "2019-01-01T10:00:00Z", "filter_by": "1D410B4BE5060972", "group_by": "queryid"}' http://127.0.0.1:9922/v0/qan/ObjectDetails/GetMetrics`


```
curl -s -X POST -d '{"period_start_from": "2019-01-01T00:00:00Z", "period_start_to": "2019-01-01T10:00:00Z", "order_by": "num_queries", "columns": ["lock_time", "sort_scan"], "group_by": "server"}' http://127.0.0.1:9922/v0/qan/GetReport -s | jq '.rows[].load'
```


```
curl -s -X POST -d '{"period_start_from": "2019-01-01T00:00:00Z", "period_start_to": "2019-01-01T10:00:00Z", "filter_by": "1D410B4BE5060972", "group_by": "queryid"}' http://127.0.0.1:9922/v0/qan/ObjectDetails/GetLabels | jq
```


## License
[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2Fpercona%2Fqan-api2.svg?type=large)](https://app.fossa.io/projects/git%2Bgithub.com%2Fpercona%2Fqan-api2?ref=badge_large)