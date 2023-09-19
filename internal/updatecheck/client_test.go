package updatecheck

import (
	"testing"

	"github.com/sourcegraph/log/logtest"
)

func TestUpdateCheckURL(t *testing.T) {
	tests := []struct {
		name           string
		env            string
		want           string
		errorLogsCount int
	}{
		{name: "default OK", env: "", want: "https://pings.sourcegraph.com/updates", errorLogsCount: 0},
		{name: "specified dotcom OK", env: "https://sourcegraph.com/", want: "https://sourcegraph.com/.api/updates", errorLogsCount: 0},
		{name: "http not OK", env: "http://sourcegraph.com/", want: "https://pings.sourcegraph.com/updates", errorLogsCount: 1},
		{name: "env OK", env: "https://fourfegraph.com/", want: "https://fourfegraph.com/.api/updates", errorLogsCount: 0},
		{name: "env no protocol not OK", env: "fourfegraph.com", want: "https://pings.sourcegraph.com/updates", errorLogsCount: 1},
		{name: "env garbage not OK", env: "garbage  foo", want: "https://pings.sourcegraph.com/updates", errorLogsCount: 1},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Setenv("UPDATE_CHECK_BASE_URL", test.env)
			logger, export := logtest.Captured(t)

			if got, expected := updateCheckURL(logger), test.want; got != expected {
				t.Errorf("Got %q, expected %q", got, expected)
			}

			if got, expected := len(export()), test.errorLogsCount; got != expected {
				t.Errorf("Got %d, expected %d log messages", got, expected)
			}
		})
	}
}

func TestParseRedisInfo(t *testing.T) {
	info, err := parseRedisInfo([]byte(redisInfoCommand))
	if err != nil {
		t.Errorf("Expected no error, got %s", err)
	}

	if got, expected := info["redis_version"], "6.0.9"; got != expected {
		t.Errorf("Got %q expected %q", got, expected)
	}
	if got, expected := info["used_memory"], "3318848"; got != expected {
		t.Errorf("Got %q expected %q", got, expected)
	}
	if got, expected := info["db0"], "keys=46,expires=17,avg_ttl=1506125315"; got != expected {
		t.Errorf("Got %q expected %q", got, expected)
	}
}

// output of running the INFO command in redis-cli
const redisInfoCommand = `# Server
redis_version:6.0.9
redis_git_sha1:00000000
redis_git_dirty:0
redis_build_id:26c3229b35eb3beb
redis_mode:standalone
os:Darwin 19.6.0 x86_64
arch_bits:64
multiplexing_api:kqueue
atomicvar_api:atomic-builtin
gcc_version:4.2.1
process_id:816
run_id:43cad29d3738a66c81072ac142da629c5be2cc1b
tcp_port:6379
uptime_in_seconds:1717
uptime_in_days:0
hz:10
configured_hz:10
lru_clock:10651157
executable:/usr/local/opt/redis/bin/redis-server
config_file:/usr/local/etc/redis.conf
io_threads_active:0

# Clients
connected_clients:6
client_recent_max_input_buffer:16
client_recent_max_output_buffer:0
blocked_clients:0
tracking_clients:0
clients_in_timeout_table:0

# Memory
used_memory:3318848
used_memory_human:3.17M
used_memory_rss:8097792
used_memory_rss_human:7.72M
used_memory_peak:3646448
used_memory_peak_human:3.48M
used_memory_peak_perc:91.02%
used_memory_overhead:1126256
used_memory_startup:1018096
used_memory_dataset:2192592
used_memory_dataset_perc:95.30%
allocator_allocated:3251760
allocator_active:8064000
allocator_resident:8064000
total_system_memory:34359738368
total_system_memory_human:32.00G
used_memory_lua:33792
used_memory_lua_human:33.00K
used_memory_scripts:440
used_memory_scripts_human:440B
number_of_cached_scripts:1
maxmemory:0
maxmemory_human:0B
maxmemory_policy:noeviction
allocator_frag_ratio:2.48
allocator_frag_bytes:4812240
allocator_rss_ratio:1.00
allocator_rss_bytes:0
rss_overhead_ratio:1.00
rss_overhead_bytes:33792
mem_fragmentation_ratio:2.49
mem_fragmentation_bytes:4846032
mem_not_counted_for_evict:0
mem_replication_backlog:0
mem_clients_slaves:0
mem_clients_normal:104704
mem_aof_buffer:0
mem_allocator:libc
active_defrag_running:0
lazyfree_pending_objects:0

# Persistence
loading:0
rdb_changes_since_last_save:0
rdb_bgsave_in_progress:0
rdb_last_save_time:1604486402
rdb_last_bgsave_status:ok
rdb_last_bgsave_time_sec:0
rdb_current_bgsave_time_sec:-1
rdb_last_cow_size:0
aof_enabled:0
aof_rewrite_in_progress:0
aof_rewrite_scheduled:0
aof_last_rewrite_time_sec:0
aof_current_rewrite_time_sec:-1
aof_last_bgrewrite_status:ok
aof_last_write_status:ok
aof_last_cow_size:0
module_fork_in_progress:0
module_fork_last_cow_size:0

# Stats
total_connections_received:35
total_commands_processed:797
instantaneous_ops_per_sec:0
total_net_input_bytes:186914
total_net_output_bytes:480700
instantaneous_input_kbps:0.01
instantaneous_output_kbps:9.79
rejected_connections:0
sync_full:0
sync_partial_ok:0
sync_partial_err:0
expired_keys:277
expired_stale_perc:0.00
expired_time_cap_reached_count:0
expire_cycle_cpu_milliseconds:61
evicted_keys:0
keyspace_hits:26
keyspace_misses:22
pubsub_channels:0
pubsub_patterns:0
latest_fork_usec:1499
migrate_cached_sockets:0
slave_expires_tracked_keys:0
active_defrag_hits:0
active_defrag_misses:0
active_defrag_key_hits:0
active_defrag_key_misses:0
tracking_total_keys:0
tracking_total_items:0
tracking_total_prefixes:0
unexpected_error_replies:0
total_reads_processed:804
total_writes_processed:773
io_threaded_reads_processed:0
io_threaded_writes_processed:0

# Replication
role:master
connected_slaves:0
master_replid:0c14ea8fafb5af311c5677d1fbe599b01843f547
master_replid2:0000000000000000000000000000000000000000
master_repl_offset:0
second_repl_offset:-1
repl_backlog_active:0
repl_backlog_size:1048576
repl_backlog_first_byte_offset:0
repl_backlog_histlen:0

# CPU
used_cpu_sys:0.639901
used_cpu_user:0.532358
used_cpu_sys_children:0.077478
used_cpu_user_children:0.135344

# Modules

# Cluster
cluster_enabled:0

# Keyspace
db0:keys=46,expires=17,avg_ttl=1506125315`
