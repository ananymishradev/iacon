package causes

const (
	RedisUnavailable   = "redis_unavailable"
	PgBouncerDown      = "pgbouncer_down"
	PostgresDown       = "postgres_down"
	DatabaseDiskFull   = "database_disk_full"
	PoolSaturated      = "pool_saturated"
	AppDown            = "app_down"
	AppInternalErrors  = "app_internal_errors"
	MonitoringDegraded = "monitoring_degraded"
	Unknown            = "unknown"
)

var all = []string{
	RedisUnavailable, PgBouncerDown, PostgresDown, DatabaseDiskFull,
	PoolSaturated, AppDown, AppInternalErrors, MonitoringDegraded, Unknown,
}

func All() []string { return append([]string(nil), all...) }

func Valid(value string) bool {
	for _, cause := range all {
		if value == cause {
			return true
		}
	}
	return false
}
