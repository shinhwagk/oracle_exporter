V$ACTIVE_SESSION_HISTORY
V$ENQUEUE_STAT
V$EVENT_HISTOGRAM
V$EVENTMETRIC
V$FILE_HISTOGRAM
V$FILESTAT
V$IOSTAT_FILE
V$IOSTAT_FUNCTION
V$LATCH *
V$ROLLSTAT
V$SEGMENT_STATISTICS
V$SEGSTAT
V$SEGSTAT_NAME
V$SERVICE_EVENT
V$SERVICE_WAIT_CLASS
V$SESSION
V$SESSION_EVENT      # session 
# V$SESSION_WAIT       # session current or last class event 
V$SESSION_WAIT_CLASS # session class level statistics, WAIT_CLASS_ID in v$event_name, 
V$SESSION_WAIT_HISTORY
V$SESSTAT            # V$STATNAME
V$SESS_TIME_MODEL
V$SYSSTAT            # V$STATNAME
V$SYSTEM_EVENT       # class event level statistics
V$SYSTEM_WAIT_CLASS # class level statistics, WAIT_CLASS_ID in v$event_name
V$SYS_TIME_MODEL
V$TEMP_HISTOGRAM
V$WAITCLASSMETRIC   # class level statistics, most recent 60-second statistics
V$WAITCLASSMETRIC_HISTORY
V$WAITSTAT #block contention statistics
V$TEMPSEG_USAGE
V$LOCK
V$SESSION_LONGOPS
V$SQL
V$SQL_PLAN
V$SQL_MONITOR
V$SQL_PLAN_MONITOR
### NO.1: examine load
redo size 
session logical reads 
db block changes 
physical reads 
physical read total bytes 
physical writes 
physical write total bytes 
parse count (total) 
parse count (hard) 
user calls 
DB time
SQL ordered by Parse Calls
### NO.2 event
