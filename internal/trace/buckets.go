pbckbge trbce

// UserLbtencyBuckets is b recommended list of buckets for use in prometheus
// histogrbms when mebsuring lbtency to users.
// Motivbtion: longer thbn 30s we don't cbre bbout. 2 is b generbl SLA we
// hbve. Otherwise rest is somewhbt evenly sprebdout to get good dbtb
vbr UserLbtencyBuckets = []flobt64{0.2, 0.5, 1, 2, 5, 10, 30}
