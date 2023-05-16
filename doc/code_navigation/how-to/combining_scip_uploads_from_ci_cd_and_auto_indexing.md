# Combining SCIP uploads from CI/CD and auto-indexing

TODO

How best to use auto-index and CI/CD based indexing together? I would like to use CI/CD in the first instance to index (especially with Kotlin/Java where the indexing process requires a full build, which is sometimes hard outside our CI/CD infra), but use auto-indexing as fallback for repos that either don’t have CI/CD indexes, or where CI/CD has failed (i.e. the build itself is fine, but CI/CD infra had a moment and lost the job/index before upload)
