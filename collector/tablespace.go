package collector

import (
	"database/sql"

	"github.com/prometheus/client_golang/prometheus"
)

type tablespaceCollector struct {
	descs [3]*prometheus.Desc
}

func init() {
	registerCollector("tablespace", cHour, defaultEnabled, NewTabalespaceCollector)
}

// NewTabalespaceCollector returns a new Collector exposing session activity statistics.
func NewTabalespaceCollector() (Collector, error) {
	descs := [3]*prometheus.Desc{
		newDesc("tablespace", "alloc_bytes", "Generic counter metric of tablespaces bytes in Oracle.", []string{"tablespace"}, nil),
		newDesc("tablespace", "max_bytes", "Generic counter metric of tablespaces bytes in Oracle.", []string{"tablespace"}, nil),
		newDesc("tablespace", "alloc_free_bytes", "Generic counter metric of tablespaces bytes in Oracle.", []string{"tablespace"}, nil),
	}
	return &tablespaceCollector{descs}, nil
}

func (c *tablespaceCollector) Update(db *sql.DB, ch chan<- prometheus.Metric) error {
	rows, err := db.Query(tablespaceSQL)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var tablespaceName string
		var bytes, maxBytes, bytesFree float64

		if err := rows.Scan(&tablespaceName, &bytesFree, &bytes, &maxBytes); err != nil {
			return err
		}

		ch <- prometheus.MustNewConstMetric(c.descs[0], prometheus.GaugeValue, bytes, tablespaceName)
		ch <- prometheus.MustNewConstMetric(c.descs[1], prometheus.GaugeValue, maxBytes, tablespaceName)
		ch <- prometheus.MustNewConstMetric(c.descs[2], prometheus.GaugeValue, bytesFree, tablespaceName)
	}
	return nil
}

const tablespaceSQL = `
SELECT ddf.tablespace_name, nvl(dfs.bytes, 0) free, ddf.bytes, ddf.MAXBYTES
  FROM (select tablespace_name, sum(DECODE(MAXBYTES, 0, bytes, MAXBYTES)) maxbytes, sum(bytes) bytes
          from dba_data_files
         group by tablespace_name) ddf,
       dba_tablespaces dt,
       (select tablespace_name, sum(bytes) bytes
          from dba_free_space
         group by tablespace_name) dfs
 WHERE ddf.tablespace_name = dt.tablespace_name
   and dt.tablespace_name = dfs.tablespace_name(+)`

// SELECT Z.name,
//   dt.status,
//   dt.contents,
//   dt.extent_management,
//   Z.bytes,
//   Z.max_bytes,
//   Z.free_bytes
// FROM
//   (SELECT X.name             AS name,
//     SUM(NVL(X.free_bytes,0)) AS free_bytes,
//     SUM(X.bytes)             AS bytes,
//     SUM(X.max_bytes)         AS max_bytes
//   FROM
//     (SELECT ddf.tablespace_name AS name,
//       ddf.status                AS status,
//       ddf.bytes                 AS bytes,
//       SUM(dfs.bytes)            AS free_bytes,
//       DECODE(ddf.maxbytes, 0, ddf.bytes, ddf.maxbytes) AS max_bytes
//     FROM dba_data_files ddf, dba_tablespaces dt, dba_free_space dfs
//     WHERE ddf.tablespace_name = dt.tablespace_name
//     AND ddf.file_id           = dfs.file_id(+)
//     GROUP BY ddf.tablespace_name,
//       ddf.file_name,
//       ddf.status,
//       ddf.bytes,
//       ddf.maxbytes
//     ) X
//   GROUP BY X.name
//   UNION ALL
//   SELECT Y.name              AS name,
//     MAX(NVL(Y.free_bytes,0)) AS free_bytes,
//     SUM(Y.bytes)             AS bytes,
//     SUM(Y.max_bytes)         AS max_bytes
//   FROM
//     (SELECT dtf.tablespace_name AS name,
//       dtf.status                AS status,
//       dtf.bytes                 AS bytes,
//       (SELECT ((f.total_blocks - s.tot_used_blocks)*vp.value)
//       FROM
//         (SELECT tablespace_name,
//           SUM(used_blocks) tot_used_blocks
//         FROM gv$sort_segment
//         WHERE tablespace_name!='DUMMY'
//         GROUP BY tablespace_name
//         ) s,
//         (SELECT tablespace_name,
//           SUM(blocks) total_blocks
//         FROM dba_temp_files
//         WHERE tablespace_name !='DUMMY'
//         GROUP BY tablespace_name
//         ) f,
//         (SELECT value FROM v$parameter WHERE name = 'db_block_size'
//         ) vp
//       WHERE f.tablespace_name=s.tablespace_name
//       AND f.tablespace_name  = dtf.tablespace_name
//       ) AS free_bytes,
//       CASE
//         WHEN dtf.maxbytes = 0
//         THEN dtf.bytes
//         ELSE dtf.maxbytes
//       END AS max_bytes
//     FROM dba_temp_files dtf
//     ) Y
//   GROUP BY Y.name
//   ) Z,
//   dba_tablespaces dt
// WHERE Z.name = dt.tablespace_name`
