SELECT
  CASE name
    WHEN 'parse count (total)' THEN 'parse_total'
    WHEN 'execute count'       THEN 'execute_total'
    WHEN 'user commits'        THEN 'commit_total'
    WHEN 'user rollbacks'      THEN 'rollback_total'
  END name,
  value
FROM v$sysstat
WHERE name IN ('parse count (total)', 'execute count', 'user commits', 'user rollbacks')