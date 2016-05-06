TESTTMP = ./test-tmp

# VANILLA REDIS CONF
define VANILLA_CONF
daemonize yes
port 6379
dir .
pidfile redis_vanilla.pid
logfile redis_vanilla.log
save ""
appendonly no
endef

# CLUSTER REDIS NODES
define NODE1_CONF
daemonize yes
port 7000
dir .
pidfile redis_cluster_node1.pid
logfile redis_cluster_node1.log
cluster-node-timeout 5000
save ""
appendonly no
cluster-enabled yes
cluster-config-file redis_cluster_node1.conf
endef

define NODE2_CONF
daemonize yes
port 7001
dir .
pidfile redis_cluster_node2.pid
logfile redis_cluster_node2.log
cluster-node-timeout 5000
save ""
appendonly no
cluster-enabled yes
cluster-config-file redis_cluster_node2.conf
endef

# SENTINEL REDIS NODES
define SENTINEL_CONF
daemonize yes
port 28000
dir .
pidfile redis_sentinel.pid
logfile redis_sentinel.log
sentinel monitor test 127.0.0.1 8000 2
sentinel down-after-milliseconds test 60000
sentinel failover-timeout test 180000
sentinel parallel-syncs test 1
endef

define SENTINEL_NODE1_CONF
daemonize yes
port 8000
dir .
pidfile redis_sentinel_node1.pid
logfile redis_sentinel_node1.log
save ""
dbfilename redis_sentinel_node1.rdb
appendonly no
endef

define SENTINEL_NODE2_CONF
daemonize yes
port 8001
dir .
pidfile redis_sentinel_node2.pid
logfile redis_sentinel_node2.log
save ""
dbfilename redis_sentinel_node1.rdb
appendonly no
endef

export VANILLA_CONF
export NODE1_CONF
export NODE2_CONF

export SENTINEL_CONF
export SENTINEL_NODE1_CONF
export SENTINEL_NODE2_CONF

.ONESHELL:
start: cleanup
	mkdir -p $(TESTTMP)
	cd $(TESTTMP)
	echo "$$VANILLA_CONF" | redis-server -
	echo "$$NODE1_CONF" | redis-server -
	echo "$$NODE2_CONF" | redis-server -
	echo "$$SENTINEL_CONF" > sentinel.conf
	redis-sentinel sentinel.conf
	echo "$$SENTINEL_NODE1_CONF" | redis-server -
	echo "$$SENTINEL_NODE2_CONF" | redis-server -
	sleep 1
	redis-cli -p 7000 cluster meet 127.0.0.1 7001
	redis-cli -p 7000 cluster addslots $$(seq 0 8191)
	redis-cli -p 7001 cluster addslots $$(seq 8192 16383)
	redis-cli -p 8001 slaveof 127.0.0.1 8000

cleanup:
	rm -rf $(TESTTMP)

stop:
	kill `cat $(TESTTMP)/redis_vanilla.pid` || true
	kill `cat $(TESTTMP)/redis_cluster_node1.pid` || true
	kill `cat $(TESTTMP)/redis_cluster_node2.pid` || true
	kill `cat $(TESTTMP)/redis_sentinel.pid` || true
	kill `cat $(TESTTMP)/redis_sentinel_node1.pid` || true
	kill `cat $(TESTTMP)/redis_sentinel_node2.pid` || true
	make cleanup
