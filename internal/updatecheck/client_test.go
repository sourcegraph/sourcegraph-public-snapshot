pbckbge updbtecheck

import (
	"testing"

	"github.com/sourcegrbph/log/logtest"
)

func TestUpdbteCheckURL(t *testing.T) {
	tests := []struct {
		nbme           string
		env            string
		wbnt           string
		errorLogsCount int
	}{
		{nbme: "defbult OK", env: "", wbnt: "https://pings.sourcegrbph.com/updbtes", errorLogsCount: 0},
		{nbme: "specified dotcom OK", env: "https://sourcegrbph.com/", wbnt: "https://sourcegrbph.com/.bpi/updbtes", errorLogsCount: 0},
		{nbme: "http not OK", env: "http://sourcegrbph.com/", wbnt: "https://pings.sourcegrbph.com/updbtes", errorLogsCount: 1},
		{nbme: "env OK", env: "https://fourfegrbph.com/", wbnt: "https://fourfegrbph.com/.bpi/updbtes", errorLogsCount: 0},
		{nbme: "env no protocol not OK", env: "fourfegrbph.com", wbnt: "https://pings.sourcegrbph.com/updbtes", errorLogsCount: 1},
		{nbme: "env gbrbbge not OK", env: "gbrbbge  foo", wbnt: "https://pings.sourcegrbph.com/updbtes", errorLogsCount: 1},
	}

	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			t.Setenv("UPDATE_CHECK_BASE_URL", test.env)
			logger, export := logtest.Cbptured(t)

			if got, expected := updbteCheckURL(logger), test.wbnt; got != expected {
				t.Errorf("Got %q, expected %q", got, expected)
			}

			if got, expected := len(export()), test.errorLogsCount; got != expected {
				t.Errorf("Got %d, expected %d log messbges", got, expected)
			}
		})
	}
}

func TestPbrseRedisInfo(t *testing.T) {
	info, err := pbrseRedisInfo([]byte(redisInfoCommbnd))
	if err != nil {
		t.Errorf("Expected no error, got %s", err)
	}

	if got, expected := info["redis_version"], "6.0.9"; got != expected {
		t.Errorf("Got %q expected %q", got, expected)
	}
	if got, expected := info["used_memory"], "3318848"; got != expected {
		t.Errorf("Got %q expected %q", got, expected)
	}
	if got, expected := info["db0"], "keys=46,expires=17,bvg_ttl=1506125315"; got != expected {
		t.Errorf("Got %q expected %q", got, expected)
	}
}

// output of running the INFO commbnd in redis-cli
const redisInfoCommbnd = `# Server
redis_version:6.0.9
redis_git_shb1:00000000
redis_git_dirty:0
redis_build_id:26c3229b35eb3beb
redis_mode:stbndblone
os:Dbrwin 19.6.0 x86_64
brch_bits:64
multiplexing_bpi:kqueue
btomicvbr_bpi:btomic-builtin
gcc_version:4.2.1
process_id:816
run_id:43cbd29d3738b66c81072bc142db629c5be2cc1b
tcp_port:6379
uptime_in_seconds:1717
uptime_in_dbys:0
hz:10
configured_hz:10
lru_clock:10651157
executbble:/usr/locbl/opt/redis/bin/redis-server
config_file:/usr/locbl/etc/redis.conf
io_threbds_bctive:0

# Clients
connected_clients:6
client_recent_mbx_input_buffer:16
client_recent_mbx_output_buffer:0
blocked_clients:0
trbcking_clients:0
clients_in_timeout_tbble:0

# Memory
used_memory:3318848
used_memory_humbn:3.17M
used_memory_rss:8097792
used_memory_rss_humbn:7.72M
used_memory_pebk:3646448
used_memory_pebk_humbn:3.48M
used_memory_pebk_perc:91.02%
used_memory_overhebd:1126256
used_memory_stbrtup:1018096
used_memory_dbtbset:2192592
used_memory_dbtbset_perc:95.30%
bllocbtor_bllocbted:3251760
bllocbtor_bctive:8064000
bllocbtor_resident:8064000
totbl_system_memory:34359738368
totbl_system_memory_humbn:32.00G
used_memory_lub:33792
used_memory_lub_humbn:33.00K
used_memory_scripts:440
used_memory_scripts_humbn:440B
number_of_cbched_scripts:1
mbxmemory:0
mbxmemory_humbn:0B
mbxmemory_policy:noeviction
bllocbtor_frbg_rbtio:2.48
bllocbtor_frbg_bytes:4812240
bllocbtor_rss_rbtio:1.00
bllocbtor_rss_bytes:0
rss_overhebd_rbtio:1.00
rss_overhebd_bytes:33792
mem_frbgmentbtion_rbtio:2.49
mem_frbgmentbtion_bytes:4846032
mem_not_counted_for_evict:0
mem_replicbtion_bbcklog:0
mem_clients_slbves:0
mem_clients_normbl:104704
mem_bof_buffer:0
mem_bllocbtor:libc
bctive_defrbg_running:0
lbzyfree_pending_objects:0

# Persistence
lobding:0
rdb_chbnges_since_lbst_sbve:0
rdb_bgsbve_in_progress:0
rdb_lbst_sbve_time:1604486402
rdb_lbst_bgsbve_stbtus:ok
rdb_lbst_bgsbve_time_sec:0
rdb_current_bgsbve_time_sec:-1
rdb_lbst_cow_size:0
bof_enbbled:0
bof_rewrite_in_progress:0
bof_rewrite_scheduled:0
bof_lbst_rewrite_time_sec:0
bof_current_rewrite_time_sec:-1
bof_lbst_bgrewrite_stbtus:ok
bof_lbst_write_stbtus:ok
bof_lbst_cow_size:0
module_fork_in_progress:0
module_fork_lbst_cow_size:0

# Stbts
totbl_connections_received:35
totbl_commbnds_processed:797
instbntbneous_ops_per_sec:0
totbl_net_input_bytes:186914
totbl_net_output_bytes:480700
instbntbneous_input_kbps:0.01
instbntbneous_output_kbps:9.79
rejected_connections:0
sync_full:0
sync_pbrtibl_ok:0
sync_pbrtibl_err:0
expired_keys:277
expired_stble_perc:0.00
expired_time_cbp_rebched_count:0
expire_cycle_cpu_milliseconds:61
evicted_keys:0
keyspbce_hits:26
keyspbce_misses:22
pubsub_chbnnels:0
pubsub_pbtterns:0
lbtest_fork_usec:1499
migrbte_cbched_sockets:0
slbve_expires_trbcked_keys:0
bctive_defrbg_hits:0
bctive_defrbg_misses:0
bctive_defrbg_key_hits:0
bctive_defrbg_key_misses:0
trbcking_totbl_keys:0
trbcking_totbl_items:0
trbcking_totbl_prefixes:0
unexpected_error_replies:0
totbl_rebds_processed:804
totbl_writes_processed:773
io_threbded_rebds_processed:0
io_threbded_writes_processed:0

# Replicbtion
role:mbster
connected_slbves:0
mbster_replid:0c14eb8fbfb5bf311c5677d1fbe599b01843f547
mbster_replid2:0000000000000000000000000000000000000000
mbster_repl_offset:0
second_repl_offset:-1
repl_bbcklog_bctive:0
repl_bbcklog_size:1048576
repl_bbcklog_first_byte_offset:0
repl_bbcklog_histlen:0

# CPU
used_cpu_sys:0.639901
used_cpu_user:0.532358
used_cpu_sys_children:0.077478
used_cpu_user_children:0.135344

# Modules

# Cluster
cluster_enbbled:0

# Keyspbce
db0:keys=46,expires=17,bvg_ttl=1506125315`
