// qan-api2
// Copyright (C) 2019 Percona LLC
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.

package models

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"sort"
	"text/template"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"

	"github.com/percona/pmm/api/qanpb"
)

const optimalAmountOfPoint = 120
const minFullTimeFrame = 2 * time.Hour

// Metrics represents methods to work with metrics.
type Metrics struct {
	db *sqlx.DB
}

// NewMetrics initialize Metrics with db instance.
func NewMetrics(db *sqlx.DB) Metrics {
	return Metrics{db: db}
}

// Get select metrics for specific queryid, hostname, etc.
func (m *Metrics) Get(ctx context.Context, periodStartFromSec, periodStartToSec int64, filter, group string,
	dimensions, labels map[string][]string) ([]M, error) {
	arg := map[string]interface{}{
		"period_start_from": periodStartFromSec,
		"period_start_to":   periodStartToSec,
	}

	tmplArgs := struct {
		PeriodStartFrom int64
		PeriodStartTo   int64
		PeriodDuration  int64
		Dimensions      map[string][]string
		Labels          map[string][]string
		DimensionVal    string
		Group           string
	}{
		periodStartFromSec,
		periodStartToSec,
		periodStartToSec - periodStartFromSec,
		dimensions,
		labels,
		filter,
		group,
	}
	var queryBuffer bytes.Buffer
	if tmpl, err := template.New("queryMetricsTmpl").Funcs(funcMap).Parse(queryMetricsTmpl); err != nil {
		log.Fatalln(err)
	} else if err = tmpl.Execute(&queryBuffer, tmplArgs); err != nil {
		log.Fatalln(err)
	}
	var results []M
	query, args, err := sqlx.Named(queryBuffer.String(), arg)
	if err != nil {
		return results, fmt.Errorf("prepare named:%v", err)
	}
	query, args, err = sqlx.In(query, args...)
	if err != nil {
		return results, fmt.Errorf("populate agruments in IN clause:%v", err)
	}
	query = m.db.Rebind(query)

	queryCtx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	rows, err := m.db.QueryxContext(queryCtx, query, args...)
	if err != nil {
		return results, fmt.Errorf("QueryxContext error:%v", err)
	}
	for rows.Next() {
		result := make(M)
		err = rows.MapScan(result)
		if err != nil {
			fmt.Printf("DimensionMetrics Scan error: %v", err)
		}
		results = append(results, result)
	}
	rows.NextResultSet()
	total := make(M)
	for rows.Next() {
		err = rows.MapScan(total)
		if err != nil {
			fmt.Printf("DimensionMetrics Scan TOTALS error: %v", err)
		}
		results = append(results, total)
	}

	return results, err
}

const queryMetricsTmpl = `
SELECT

SUM(num_queries) AS num_queries,
SUM(num_queries_with_errors) AS num_queries_with_errors,
SUM(num_queries_with_warnings) AS num_queries_with_warnings,

SUM(m_query_time_cnt) AS m_query_time_cnt,
SUM(m_query_time_sum) AS m_query_time_sum,
MIN(m_query_time_min) AS m_query_time_min,
MAX(m_query_time_max) AS m_query_time_max,
AVG(m_query_time_p99) AS m_query_time_p99,

SUM(m_lock_time_cnt) AS m_lock_time_cnt,
SUM(m_lock_time_sum) AS m_lock_time_sum,
MIN(m_lock_time_min) AS m_lock_time_min,
MAX(m_lock_time_max) AS m_lock_time_max,
AVG(m_lock_time_p99) AS m_lock_time_p99,

SUM(m_rows_sent_cnt) AS m_rows_sent_cnt,
SUM(m_rows_sent_sum) AS m_rows_sent_sum,
MIN(m_rows_sent_min) AS m_rows_sent_min,
MAX(m_rows_sent_max) AS m_rows_sent_max,
AVG(m_rows_sent_p99) AS m_rows_sent_p99,

SUM(m_rows_examined_cnt) AS m_rows_examined_cnt,
SUM(m_rows_examined_sum) AS m_rows_examined_sum,
MIN(m_rows_examined_min) AS m_rows_examined_min,
MAX(m_rows_examined_max) AS m_rows_examined_max,
AVG(m_rows_examined_p99) AS m_rows_examined_p99,

SUM(m_rows_affected_cnt) AS m_rows_affected_cnt,
SUM(m_rows_affected_sum) AS m_rows_affected_sum,
MIN(m_rows_affected_min) AS m_rows_affected_min,
MAX(m_rows_affected_max) AS m_rows_affected_max,
AVG(m_rows_affected_p99) AS m_rows_affected_p99,

SUM(m_rows_read_cnt) AS m_rows_read_cnt,
SUM(m_rows_read_sum) AS m_rows_read_sum,
MIN(m_rows_read_min) AS m_rows_read_min,
MAX(m_rows_read_max) AS m_rows_read_max,
AVG(m_rows_read_p99) AS m_rows_read_p99,

SUM(m_merge_passes_cnt) AS m_merge_passes_cnt,
SUM(m_merge_passes_sum) AS m_merge_passes_sum,
MIN(m_merge_passes_min) AS m_merge_passes_min,
MAX(m_merge_passes_max) AS m_merge_passes_max,
AVG(m_merge_passes_p99) AS m_merge_passes_p99,

SUM(m_innodb_io_r_ops_cnt) AS m_innodb_io_r_ops_cnt,
SUM(m_innodb_io_r_ops_sum) AS m_innodb_io_r_ops_sum,
MIN(m_innodb_io_r_ops_min) AS m_innodb_io_r_ops_min,
MAX(m_innodb_io_r_ops_max) AS m_innodb_io_r_ops_max,
AVG(m_innodb_io_r_ops_p99) AS m_innodb_io_r_ops_p99,

SUM(m_innodb_io_r_bytes_cnt) AS m_innodb_io_r_bytes_cnt,
SUM(m_innodb_io_r_bytes_sum) AS m_innodb_io_r_bytes_sum,
MIN(m_innodb_io_r_bytes_min) AS m_innodb_io_r_bytes_min,
MAX(m_innodb_io_r_bytes_max) AS m_innodb_io_r_bytes_max,
AVG(m_innodb_io_r_bytes_p99) AS m_innodb_io_r_bytes_p99,

SUM(m_innodb_io_r_wait_cnt) AS m_innodb_io_r_wait_cnt,
SUM(m_innodb_io_r_wait_sum) AS m_innodb_io_r_wait_sum,
MIN(m_innodb_io_r_wait_min) AS m_innodb_io_r_wait_min,
MAX(m_innodb_io_r_wait_max) AS m_innodb_io_r_wait_max,
AVG(m_innodb_io_r_wait_p99) AS m_innodb_io_r_wait_p99,

SUM(m_innodb_rec_lock_wait_cnt) AS m_innodb_rec_lock_wait_cnt,
SUM(m_innodb_rec_lock_wait_sum) AS m_innodb_rec_lock_wait_sum,
MIN(m_innodb_rec_lock_wait_min) AS m_innodb_rec_lock_wait_min,
MAX(m_innodb_rec_lock_wait_max) AS m_innodb_rec_lock_wait_max,
AVG(m_innodb_rec_lock_wait_p99) AS m_innodb_rec_lock_wait_p99,

SUM(m_innodb_queue_wait_cnt) AS m_innodb_queue_wait_cnt,
SUM(m_innodb_queue_wait_sum) AS m_innodb_queue_wait_sum,
MIN(m_innodb_queue_wait_min) AS m_innodb_queue_wait_min,
MAX(m_innodb_queue_wait_max) AS m_innodb_queue_wait_max,
AVG(m_innodb_queue_wait_p99) AS m_innodb_queue_wait_p99,

SUM(m_innodb_pages_distinct_cnt) AS m_innodb_pages_distinct_cnt,
SUM(m_innodb_pages_distinct_sum) AS m_innodb_pages_distinct_sum,
MIN(m_innodb_pages_distinct_min) AS m_innodb_pages_distinct_min,
MAX(m_innodb_pages_distinct_max) AS m_innodb_pages_distinct_max,
AVG(m_innodb_pages_distinct_p99) AS m_innodb_pages_distinct_p99,

SUM(m_query_length_cnt) AS m_query_length_cnt,
SUM(m_query_length_sum) AS m_query_length_sum,
MIN(m_query_length_min) AS m_query_length_min,
MAX(m_query_length_max) AS m_query_length_max,
AVG(m_query_length_p99) AS m_query_length_p99,

SUM(m_bytes_sent_cnt) AS m_bytes_sent_cnt,
SUM(m_bytes_sent_sum) AS m_bytes_sent_sum,
MIN(m_bytes_sent_min) AS m_bytes_sent_min,
MAX(m_bytes_sent_max) AS m_bytes_sent_max,
AVG(m_bytes_sent_p99) AS m_bytes_sent_p99,

SUM(m_tmp_tables_cnt) AS m_tmp_tables_cnt,
SUM(m_tmp_tables_sum) AS m_tmp_tables_sum,
MIN(m_tmp_tables_min) AS m_tmp_tables_min,
MAX(m_tmp_tables_max) AS m_tmp_tables_max,
AVG(m_tmp_tables_p99) AS m_tmp_tables_p99,

SUM(m_tmp_disk_tables_cnt) AS m_tmp_disk_tables_cnt,
SUM(m_tmp_disk_tables_sum) AS m_tmp_disk_tables_sum,
MIN(m_tmp_disk_tables_min) AS m_tmp_disk_tables_min,
MAX(m_tmp_disk_tables_max) AS m_tmp_disk_tables_max,
AVG(m_tmp_disk_tables_p99) AS m_tmp_disk_tables_p99,

SUM(m_tmp_table_sizes_cnt) AS m_tmp_table_sizes_cnt,
SUM(m_tmp_table_sizes_sum) AS m_tmp_table_sizes_sum,
MIN(m_tmp_table_sizes_min) AS m_tmp_table_sizes_min,
MAX(m_tmp_table_sizes_max) AS m_tmp_table_sizes_max,
AVG(m_tmp_table_sizes_p99) AS m_tmp_table_sizes_p99,

SUM(m_qc_hit_sum) AS m_qc_hit_sum,
SUM(m_full_scan_sum) AS m_full_scan_sum,
SUM(m_full_join_sum) AS m_full_join_sum,
SUM(m_tmp_table_sum) AS m_tmp_table_sum,
SUM(m_tmp_table_on_disk_sum) AS m_tmp_table_on_disk_sum,
SUM(m_filesort_sum) AS m_filesort_sum,
SUM(m_filesort_on_disk_sum) AS m_filesort_on_disk_sum,
SUM(m_select_full_range_join_sum) AS m_select_full_range_join_sum,
SUM(m_select_range_sum) AS m_select_range_sum,
SUM(m_select_range_check_sum) AS m_select_range_check_sum,
SUM(m_sort_range_sum) AS m_sort_range_sum,
SUM(m_sort_rows_sum) AS m_sort_rows_sum,
SUM(m_sort_scan_sum) AS m_sort_scan_sum,
SUM(m_no_index_used_sum) AS m_no_index_used_sum,
SUM(m_no_good_index_used_sum) AS m_no_good_index_used_sum,

SUM(m_docs_returned_cnt) AS m_docs_returned_cnt,
SUM(m_docs_returned_sum) AS m_docs_returned_sum,
MIN(m_docs_returned_min) AS m_docs_returned_min,
MAX(m_docs_returned_max) AS m_docs_returned_max,
AVG(m_docs_returned_p99) AS m_docs_returned_p99,

SUM(m_response_length_cnt) AS m_response_length_cnt,
SUM(m_response_length_sum) AS m_response_length_sum,
MIN(m_response_length_min) AS m_response_length_min,
MAX(m_response_length_max) AS m_response_length_max,
AVG(m_response_length_p99) AS m_response_length_p99,

SUM(m_docs_scanned_cnt) AS m_docs_scanned_cnt,
SUM(m_docs_scanned_sum) AS m_docs_scanned_sum,
MIN(m_docs_scanned_min) AS m_docs_scanned_min,
MAX(m_docs_scanned_max) AS m_docs_scanned_max,
AVG(m_docs_scanned_p99) AS m_docs_scanned_p99,

SUM(m_shared_blks_hit_sum) AS m_shared_blks_hit_sum,
SUM(m_shared_blks_read_sum) AS m_shared_blks_read_sum,
SUM(m_shared_blks_dirtied_sum) AS m_shared_blks_dirtied_sum,
SUM(m_shared_blks_written_sum) AS m_shared_blks_written_sum,

SUM(m_local_blks_hit_sum) AS m_local_blks_hit_sum,
SUM(m_local_blks_read_sum) AS m_local_blks_read_sum,
SUM(m_local_blks_dirtied_sum) AS m_local_blks_dirtied_sum,
SUM(m_local_blks_written_sum) AS m_local_blks_written_sum,

SUM(m_temp_blks_read_sum) AS m_temp_blks_read_sum,
SUM(m_temp_blks_written_sum) AS m_temp_blks_written_sum,
SUM(m_blk_read_time_sum) AS m_blk_read_time_sum,
SUM(m_blk_write_time_sum) AS m_blk_write_time_sum

FROM metrics
WHERE period_start >= :period_start_from AND period_start <= :period_start_to
{{ if .DimensionVal }} AND {{ .Group }} = '{{ .DimensionVal }}' {{ end }}
{{ if .Dimensions }}
    {{range $key, $vals := .Dimensions }}
        AND {{ $key }} IN ( '{{ StringsJoin $vals "', '" }}' )
    {{ end }}
{{ end }}
{{ if .Labels }}{{$i := 0}}
    AND ({{range $key, $vals := .Labels }}{{ $i = inc $i}}
        {{ if gt $i 1}} OR {{ end }} has(['{{ StringsJoin $vals "', '" }}'], labels.value[indexOf(labels.key, '{{ $key }}')])
    {{ end }})
{{ end }}
{{ if .DimensionVal }} GROUP BY {{ .Group }} {{ end }}
	WITH TOTALS;
`

const queryMetricsSparklinesTmpl = `
SELECT
intDivOrZero(toUnixTimestamp( :period_start_to ) - toUnixTimestamp(period_start), {{ .TimeFrame }}) AS point,
toDateTime(toUnixTimestamp( :period_start_to ) - (point * {{ .TimeFrame }})) AS timestamp,
{{ .TimeFrame }} AS time_frame,

SUM(m_query_time_sum) / time_frame AS load,
SUM(num_queries) / time_frame AS num_queries_per_sec,
SUM(num_queries_with_errors) / time_frame AS num_queries_with_errors_per_sec,
SUM(num_queries_with_warnings) / time_frame AS num_queries_with_warnings_per_sec,
if(SUM(m_query_time_cnt) == 0, NaN, load) AS m_query_time_sum_per_sec,
if(SUM(m_lock_time_cnt) == 0, NaN, SUM(m_lock_time_sum) / time_frame) AS m_lock_time_sum_per_sec,
if(SUM(m_rows_sent_cnt) == 0, NaN, SUM(m_rows_sent_sum) / time_frame) AS m_rows_sent_sum_per_sec,
if(SUM(m_rows_examined_cnt) == 0, NaN, SUM(m_rows_examined_sum) / time_frame) AS m_rows_examined_sum_per_sec,
if(SUM(m_rows_affected_cnt) == 0, NaN, SUM(m_rows_affected_sum) / time_frame) AS m_rows_affected_sum_per_sec,
if(SUM(m_rows_read_cnt) == 0, NaN, SUM(m_rows_read_sum) / time_frame) AS m_rows_read_sum_per_sec,
if(SUM(m_merge_passes_cnt) == 0, NaN, SUM(m_merge_passes_sum) / time_frame) AS m_merge_passes_sum_per_sec,
if(SUM(m_innodb_io_r_ops_cnt) == 0, NaN, SUM(m_innodb_io_r_ops_sum) / time_frame) AS m_innodb_io_r_ops_sum_per_sec,
if(SUM(m_innodb_io_r_bytes_cnt) == 0, NaN, SUM(m_innodb_io_r_bytes_sum) / time_frame) AS m_innodb_io_r_bytes_sum_per_sec,
if(SUM(m_innodb_io_r_wait_cnt) == 0, NaN, SUM(m_innodb_io_r_wait_sum) / time_frame) AS m_innodb_io_r_wait_sum_per_sec,
if(SUM(m_innodb_rec_lock_wait_cnt) == 0, NaN, SUM(m_innodb_rec_lock_wait_sum) / time_frame) AS m_innodb_rec_lock_wait_sum_per_sec,
if(SUM(m_innodb_queue_wait_cnt) == 0, NaN, SUM(m_innodb_queue_wait_sum) / time_frame) AS m_innodb_queue_wait_sum_per_sec,
if(SUM(m_innodb_pages_distinct_cnt) == 0, NaN, SUM(m_innodb_pages_distinct_sum) / time_frame) AS m_innodb_pages_distinct_sum_per_sec,
if(SUM(m_query_length_cnt) == 0, NaN, SUM(m_query_length_sum) / time_frame) AS m_query_length_sum_per_sec,
if(SUM(m_bytes_sent_cnt) == 0, NaN, SUM(m_bytes_sent_sum) / time_frame) AS m_bytes_sent_sum_per_sec,
if(SUM(m_tmp_tables_cnt) == 0, NaN, SUM(m_tmp_tables_sum) / time_frame) AS m_tmp_tables_sum_per_sec,
if(SUM(m_tmp_disk_tables_cnt) == 0, NaN, SUM(m_tmp_disk_tables_sum) / time_frame) AS m_tmp_disk_tables_sum_per_sec,
if(SUM(m_tmp_table_sizes_cnt) == 0, NaN, SUM(m_tmp_table_sizes_sum) / time_frame) AS m_tmp_table_sizes_sum_per_sec,
if(SUM(m_qc_hit_cnt) == 0, NaN, SUM(m_qc_hit_sum) / time_frame) AS m_qc_hit_sum_per_sec,
if(SUM(m_full_scan_cnt) == 0, NaN, SUM(m_full_scan_sum) / time_frame) AS m_full_scan_sum_per_sec,
if(SUM(m_full_join_cnt) == 0, NaN, SUM(m_full_join_sum) / time_frame) AS m_full_join_sum_per_sec,
if(SUM(m_tmp_table_cnt) == 0, NaN, SUM(m_tmp_table_sum) / time_frame) AS m_tmp_table_sum_per_sec,
if(SUM(m_tmp_table_on_disk_cnt) == 0, NaN, SUM(m_tmp_table_on_disk_sum) / time_frame) AS m_tmp_table_on_disk_sum_per_sec,
if(SUM(m_filesort_cnt) == 0, NaN, SUM(m_filesort_sum) / time_frame) AS m_filesort_sum_per_sec,
if(SUM(m_filesort_on_disk_cnt) == 0, NaN, SUM(m_filesort_on_disk_sum) / time_frame) AS m_filesort_on_disk_sum_per_sec,
if(SUM(m_select_full_range_join_cnt) == 0, NaN, SUM(m_select_full_range_join_sum) / time_frame) AS m_select_full_range_join_sum_per_sec,
if(SUM(m_select_range_cnt) == 0, NaN, SUM(m_select_range_sum) / time_frame) AS m_select_range_sum_per_sec,
if(SUM(m_select_range_check_cnt) == 0, NaN, SUM(m_select_range_check_sum) / time_frame) AS m_select_range_check_sum_per_sec,
if(SUM(m_sort_range_cnt) == 0, NaN, SUM(m_sort_range_sum) / time_frame) AS m_sort_range_sum_per_sec,
if(SUM(m_sort_rows_cnt) == 0, NaN, SUM(m_sort_rows_sum) / time_frame) AS m_sort_rows_sum_per_sec,
if(SUM(m_sort_scan_cnt) == 0, NaN, SUM(m_sort_scan_sum) / time_frame) AS m_sort_scan_sum_per_sec,
if(SUM(m_no_index_used_cnt) == 0, NaN, SUM(m_no_index_used_sum) / time_frame) AS m_no_index_used_sum_per_sec,
if(SUM(m_no_good_index_used_cnt) == 0, NaN, SUM(m_no_good_index_used_sum) / time_frame) AS m_no_good_index_used_sum_per_sec,
if(SUM(m_docs_returned_cnt) == 0, NaN, SUM(m_docs_returned_sum) / time_frame) AS m_docs_returned_sum_per_sec,
if(SUM(m_response_length_cnt) == 0, NaN, SUM(m_response_length_sum) / time_frame) AS m_response_length_sum_per_sec,
if(SUM(m_docs_scanned_cnt) == 0, NaN, SUM(m_docs_scanned_sum) / time_frame) AS m_docs_scanned_sum_per_sec,
if(SUM(m_shared_blks_hit_cnt) == 0, NaN, SUM(m_shared_blks_hit_sum) / time_frame) AS m_shared_blks_hit_sum_per_sec,
if(SUM(m_shared_blks_read_cnt) == 0, NaN, SUM(m_shared_blks_read_sum) / time_frame) AS m_shared_blks_read_sum_per_sec,
if(SUM(m_shared_blks_dirtied_cnt) == 0, NaN, SUM(m_shared_blks_dirtied_sum) / time_frame) AS m_shared_blks_dirtied_sum_per_sec,
if(SUM(m_shared_blks_written_cnt) == 0, NaN, SUM(m_shared_blks_written_sum) / time_frame) AS m_shared_blks_written_sum_per_sec,
if(SUM(m_local_blks_hit_cnt) == 0, NaN, SUM(m_local_blks_hit_sum) / time_frame) AS m_local_blks_hit_sum_per_sec,
if(SUM(m_local_blks_read_cnt) == 0, NaN, SUM(m_local_blks_read_sum) / time_frame) AS m_local_blks_read_sum_per_sec,
if(SUM(m_local_blks_dirtied_cnt) == 0, NaN, SUM(m_local_blks_dirtied_sum) / time_frame) AS m_local_blks_dirtied_sum_per_sec,
if(SUM(m_local_blks_written_cnt) == 0, NaN, SUM(m_local_blks_written_sum) / time_frame) AS m_local_blks_written_sum_per_sec,
if(SUM(m_temp_blks_read_cnt) == 0, NaN, SUM(m_temp_blks_read_sum) / time_frame) AS m_temp_blks_read_sum_per_sec,
if(SUM(m_temp_blks_written_cnt) == 0, NaN, SUM(m_temp_blks_written_sum) / time_frame) AS m_temp_blks_written_sum_per_sec,
if(SUM(m_blk_read_time_cnt) == 0, NaN, SUM(m_blk_read_time_sum) / time_frame) AS m_blk_read_time_sum_per_sec,
if(SUM(m_blk_write_time_cnt) == 0, NaN, SUM(m_blk_write_time_sum) / time_frame) AS m_blk_write_time_sum_per_sec
FROM metrics
WHERE period_start >= :period_start_from AND period_start <= :period_start_to
{{ if .DimensionVal }} AND {{ .Group }} = '{{ .DimensionVal }}' {{ end }}
{{ if .Dimensions }}
    {{range $key, $vals := .Dimensions }}
        AND {{ $key }} IN ( '{{ StringsJoin $vals "', '" }}' )
    {{ end }}
{{ end }}
{{ if .Labels }}{{$i := 0}}
    AND ({{range $key, $vals := .Labels }}{{ $i = inc $i}}
        {{ if gt $i 1}} OR {{ end }} has(['{{ StringsJoin $vals "', '" }}'], labels.value[indexOf(labels.key, '{{ $key }}')])
    {{ end }})
{{ end }}
GROUP BY point
	ORDER BY point ASC;
`

//nolint
var tmplMetricsSparklines = template.Must(template.New("queryMetricsSparklines").Funcs(funcMap).Parse(queryMetricsSparklinesTmpl))

// SelectSparklines selects datapoint for sparklines.
func (m *Metrics) SelectSparklines(ctx context.Context, periodStartFromSec, periodStartToSec int64,
	filter, group string, dimensions, labels map[string][]string) ([]*qanpb.Point, error) {

	// Align to minutes
	periodStartToSec = periodStartToSec / 60 * 60
	periodStartFromSec = periodStartFromSec / 60 * 60

	// If time range is bigger then two hour - amount of sparklines points = 120 to avoid huge data in response.
	// Otherwise amount of sparklines points is equal to minutes in in time range to not mess up calculation.
	amountOfPoints := int64(optimalAmountOfPoint)
	timePeriod := periodStartToSec - periodStartFromSec
	// reduce amount of point if period less then 2h.
	if timePeriod < int64((minFullTimeFrame).Seconds()) {
		// minimum point is 1 minute
		amountOfPoints = timePeriod / 60
	}

	// how many full minutes we can fit into given amount of points.
	minutesInPoint := (periodStartToSec - periodStartFromSec) / 60 / amountOfPoints
	// we need aditional point to show this minutes
	remainder := ((periodStartToSec - periodStartFromSec) / 60) % amountOfPoints
	amountOfPoints += remainder / minutesInPoint
	timeFrame := minutesInPoint * 60

	arg := map[string]interface{}{
		"period_start_from": periodStartFromSec,
		"period_start_to":   periodStartToSec,
	}

	tmplArgs := struct {
		PeriodStartFrom int64
		PeriodStartTo   int64
		PeriodDuration  int64
		Dimensions      map[string][]string
		Labels          map[string][]string
		DimensionVal    string
		TimeFrame       int64
		Group           string
	}{
		periodStartFromSec,
		periodStartToSec,
		periodStartToSec - periodStartFromSec,
		dimensions,
		labels,
		filter,
		timeFrame,
		group,
	}

	var results []*qanpb.Point
	var queryBuffer bytes.Buffer
	if err := tmplMetricsSparklines.Execute(&queryBuffer, tmplArgs); err != nil {
		return nil, errors.Wrap(err, "cannot execute tmplMetricsSparklines")
	}
	query, args, err := sqlx.Named(queryBuffer.String(), arg)
	if err != nil {
		return nil, errors.Wrap(err, "prepare named")
	}
	query, args, err = sqlx.In(query, args...)
	if err != nil {
		return nil, errors.Wrap(err, "populate agruments in IN clause")
	}
	query = m.db.Rebind(query)

	queryCtx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	rows, err := m.db.QueryxContext(queryCtx, query, args...)
	if err != nil {
		return nil, errors.Wrap(err, "metrics sparklines query")
	}
	resultsWithGaps := map[uint32]*qanpb.Point{}
	for rows.Next() {
		p := qanpb.Point{}
		res := getPointFieldsList(&p, sparklinePointAllFields)
		err = rows.Scan(res...)
		if err != nil {
			return nil, errors.Wrap(err, "DimensionReport scan error")
		}
		resultsWithGaps[p.Point] = &p
	}

	// fill in gaps in time series.
	for pointN := uint32(0); int64(pointN) < amountOfPoints; pointN++ {
		p, ok := resultsWithGaps[pointN]
		if !ok {
			p = &qanpb.Point{}
			p.Point = pointN
			p.TimeFrame = uint32(timeFrame)
			timeShift := timeFrame * int64(pointN)
			ts := periodStartToSec - timeShift
			p.Timestamp = time.Unix(ts, 0).UTC().Format(time.RFC3339)
		}
		results = append(results, p)
	}

	return results, err
}

const queryExampleTmpl = `
SELECT schema AS schema, tables, service_id, service_type, example, toUInt8(example_format) AS example_format,
       is_truncated, toUInt8(example_type) AS example_type, example_metrics
  FROM metrics
 WHERE period_start >= :period_start_from AND period_start <= :period_start_to
 {{ if .DimensionVal }} AND {{ .Group }} = :filter {{ end }}
 {{ if .Dimensions }}
    {{range $key, $vals := .Dimensions }}
        AND {{ $key }} IN ( '{{ StringsJoin $vals "', '" }}' )
    {{ end }}
 {{ end }}
 {{ if .Labels }}{{$i := 0}}
    AND ({{range $key, $vals := .Labels }}{{ $i = inc $i}}
        {{ if gt $i 1}} OR {{ end }} has(['{{ StringsJoin $vals "', '" }}'], labels.value[indexOf(labels.key, '{{ $key }}')])
        {{ end }})
 {{ end }}
 LIMIT :limit
`

//nolint
var tmplQueryExample = template.Must(template.New("queryExampleTmpl").Funcs(funcMap).Parse(queryExampleTmpl))

// SelectQueryExamples selects query examples and related stuff for given time range.
func (m *Metrics) SelectQueryExamples(ctx context.Context, periodStartFrom, periodStartTo time.Time, filter,
	group string, limit uint32, dimensions, labels map[string][]string) (*qanpb.QueryExampleReply, error) {
	arg := map[string]interface{}{
		"filter":            filter,
		"group":             group,
		"period_start_to":   periodStartTo,
		"period_start_from": periodStartFrom,
		"limit":             limit,
	}

	tmplArgs := struct {
		Dimensions   map[string][]string
		Labels       map[string][]string
		DimensionVal string
		Group        string
	}{
		dimensions,
		labels,
		filter,
		group,
	}

	var queryBuffer bytes.Buffer
	if err := tmplQueryExample.Execute(&queryBuffer, tmplArgs); err != nil {
		return nil, errors.Wrap(err, "cannot execute queryExampleTmpl")
	}
	query, queryArgs, err := sqlx.Named(queryBuffer.String(), arg)
	if err != nil {
		return nil, errors.Wrap(err, "prepare named")
	}
	query = m.db.Rebind(query)
	rows, err := m.db.QueryContext(ctx, query, queryArgs...)
	if err != nil {
		return nil, errors.Wrap(err, "cannot select object details labels")
	}
	defer rows.Close() //nolint:errcheck

	res := qanpb.QueryExampleReply{}
	for rows.Next() {
		var row qanpb.QueryExample
		err = rows.Scan(
			&row.Schema,
			&row.Tables,
			&row.ServiceId,
			&row.ServiceType,
			&row.Example,
			&row.ExampleFormat,
			&row.IsTruncated,
			&row.ExampleType,
			&row.ExampleMetrics,
		)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan query example for object details")
		}
		res.QueryExamples = append(res.QueryExamples, &row)
	}

	return &res, nil
}

const queryObjectDetailsLabelsTmpl = `
	SELECT
	service_name,
	database,
	schema,
	username,
	client_host,
	replication_set,
	cluster,
	service_type,
	service_id,
	environment,
	az,
	region,
	node_model,
	node_id,
	node_name,
	node_type,
	machine_id,
	container_name,
	container_id,
	agent_id,
	agent_type,
	labels.key AS lkey,
	labels.value AS lvalue
	FROM metrics
	LEFT ARRAY JOIN labels
	WHERE period_start >= :period_start_from AND period_start <= :period_start_to
	{{ if index . "filter" }} AND {{ index . "group" }} = :filter {{ end }}
	ORDER BY service_name, database, schema, username, client_host, replication_set, cluster, service_type, service_id,
			 environment, az, region, node_model, node_id, node_name, node_type, machine_id, container_name, container_id,
			 agent_id, agent_type, labels.key, labels.value
`

//nolint
var tmplObjectDetailsLabels = template.Must(template.New("queryObjectDetailsLabelsTmpl").Funcs(funcMap).Parse(queryObjectDetailsLabelsTmpl))

type queryRowsLabels struct {
	ServiceName    string
	Database       string
	Schema         string
	Username       string
	ClientHost     string
	ReplicationSet string
	Cluster        string
	ServiceType    string
	ServiceID      string
	Environment    string
	AZ             string
	Region         string
	NodeModel      string
	NodeID         string
	NodeName       string
	NodeType       string
	MachineID      string
	ContainerName  string
	ContainerID    string
	AgentID        string
	AgentType      string
	LabelKey       string
	LabelValue     string
}

// SelectObjectDetailsLabels selects object details labels for given time range and object.
func (m *Metrics) SelectObjectDetailsLabels(ctx context.Context, periodStartFrom, periodStartTo time.Time, filter,
	group string) (*qanpb.ObjectDetailsLabelsReply, error) {
	arg := map[string]interface{}{
		"filter":            filter,
		"group":             group,
		"period_start_to":   periodStartTo,
		"period_start_from": periodStartFrom,
	}

	var queryBuffer bytes.Buffer
	if err := tmplObjectDetailsLabels.Execute(&queryBuffer, arg); err != nil {
		return nil, errors.Wrap(err, "cannot execute tmplObjectDetailsLabels")
	}
	res := qanpb.ObjectDetailsLabelsReply{}

	query, queryArgs, err := sqlx.Named(queryBuffer.String(), arg)
	if err != nil {
		return nil, errors.Wrap(err, "prepare named")
	}
	query = m.db.Rebind(query)
	rows, err := m.db.QueryContext(ctx, query, queryArgs...)
	if err != nil {
		return nil, errors.Wrap(err, "cannot select object details labels")
	}
	defer rows.Close() //nolint:errcheck

	labels := map[string]map[string]struct{}{}
	labels["service_name"] = map[string]struct{}{}
	labels["database"] = map[string]struct{}{}
	labels["schema"] = map[string]struct{}{}
	labels["client_host"] = map[string]struct{}{}
	labels["username"] = map[string]struct{}{}
	labels["replication_set"] = map[string]struct{}{}
	labels["cluster"] = map[string]struct{}{}
	labels["service_type"] = map[string]struct{}{}
	labels["service_id"] = map[string]struct{}{}
	labels["environment"] = map[string]struct{}{}
	labels["az"] = map[string]struct{}{}
	labels["region"] = map[string]struct{}{}
	labels["node_model"] = map[string]struct{}{}
	labels["node_id"] = map[string]struct{}{}
	labels["node_name"] = map[string]struct{}{}
	labels["node_type"] = map[string]struct{}{}
	labels["machine_id"] = map[string]struct{}{}
	labels["container_name"] = map[string]struct{}{}
	labels["container_id"] = map[string]struct{}{}
	labels["agent_id"] = map[string]struct{}{}
	labels["agent_type"] = map[string]struct{}{}

	for rows.Next() {
		var row queryRowsLabels
		err = rows.Scan(
			&row.ServiceName,
			&row.Database,
			&row.Schema,
			&row.Username,
			&row.ClientHost,
			&row.ReplicationSet,
			&row.Cluster,
			&row.ServiceType,
			&row.ServiceID,
			&row.Environment,
			&row.AZ,
			&row.Region,
			&row.NodeModel,
			&row.NodeID,
			&row.NodeName,
			&row.NodeType,
			&row.MachineID,
			&row.ContainerName,
			&row.ContainerID,
			&row.AgentID,
			&row.AgentType,
			&row.LabelKey,
			&row.LabelValue,
		)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan labels for object details")
		}
		// convert rows to array of unique label keys - values.
		labels["service_name"][row.ServiceName] = struct{}{}
		labels["database"][row.Database] = struct{}{}
		labels["schema"][row.Schema] = struct{}{}
		labels["username"][row.Username] = struct{}{}
		labels["client_host"][row.ClientHost] = struct{}{}
		labels["replication_set"][row.ReplicationSet] = struct{}{}
		labels["cluster"][row.Cluster] = struct{}{}
		labels["service_type"][row.ServiceType] = struct{}{}
		labels["service_id"][row.ServiceID] = struct{}{}
		labels["environment"][row.Environment] = struct{}{}
		labels["az"][row.AZ] = struct{}{}
		labels["region"][row.Region] = struct{}{}
		labels["node_model"][row.NodeModel] = struct{}{}
		labels["node_id"][row.NodeID] = struct{}{}
		labels["node_name"][row.NodeName] = struct{}{}
		labels["node_type"][row.NodeType] = struct{}{}
		labels["machine_id"][row.MachineID] = struct{}{}
		labels["container_name"][row.ContainerName] = struct{}{}
		labels["container_id"][row.ContainerID] = struct{}{}
		labels["agent_id"][row.AgentID] = struct{}{}
		labels["agent_type"][row.AgentType] = struct{}{}
		if row.LabelKey != "" {
			if labels[row.LabelKey] == nil {
				labels[row.LabelKey] = map[string]struct{}{}
			}
			labels[row.LabelKey][row.LabelValue] = struct{}{}
		}
	}
	if err = rows.Err(); err != nil {
		return nil, errors.Wrap(err, "failed to select labels dimensions")
	}

	res.Labels = map[string]*qanpb.ListLabelValues{}
	// rearrange labels into gRPC response structure.
	for key, values := range labels {
		if res.Labels[key] == nil {
			res.Labels[key] = &qanpb.ListLabelValues{
				Values: []string{},
			}
		}
		for value := range values {
			res.Labels[key].Values = append(res.Labels[key].Values, value)
		}
		sort.Strings(res.Labels[key].Values)
	}

	return &res, nil
}
