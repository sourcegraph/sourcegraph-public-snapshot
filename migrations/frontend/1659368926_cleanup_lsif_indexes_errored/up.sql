UPDATE lsif_indexes SET state = 'failed' WHERE state = 'errored' AND num_failures > 0;
