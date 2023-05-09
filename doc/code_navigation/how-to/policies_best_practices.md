# Code intelligence policies best practices

This guide gives an overview of best practices when defining [auto-index scheduling](configure_auto_indexing.md) and [data retention](configure_data_retention.md) policies as it relates to resource usage (particular the disk requirement for Postgres).

**Auto-index scheduling policies** define the _cadence_ at which particular repositories will be subject to have indexing jobs scheduled (depending on the repository's configuration and contents). These policies define when a commit of a repository should be marked as an auto-index job candidate. 

**Data retention policies** define the set of indexes which are _interesting_ to a particular Sourcegraph instance's users. This is generally not *all* commits (or even all repositories), and can change drastically from company-to-company. SCIP indexes for the head of a repository's default branch are useful to explore what is most likely the current view of the development branch. These are safe from expiration by default and do not require an explicit policy to be defined by the user.

SCIP indexes for historic commits can be useful to retain in addition to the current development target (retained implicitly). SCIP indexes associated with git tags matching `v*` (for example) may be useful to index and retain indefinitely (or for some relevant length of time). These indexes are _more_ likely to be used by other projects, running in production, or used externally than any arbitrary commit of the repository. Branches with a particular format may also have significance indicating its indexes should not be garbage collected. Policies should be created in such a manner that the commits most useful to _your_ engineers are covered. This may mean defining both an auto-indexing and data retention policy, or ensuring that SCIP indexes uploaded externally (via build system) are covered by a data retention policy.

However, _usefulness_ vs _cost_ can be a delicate balancing act. If all SCIP indexes uploaded to an instance were retained forever, then disk requirements for Postgres would grow continuously. This is not sustainable for any engineering team with a non-zero commit velocity.

> WARNING: if your engineering team has zero velocity please contact Sourcegraph support for guidance.

Both of these types of policies (as well as any SCIP uploads integrated into a build system) influence the usage of disk in the `codeintel-db` schema. There are two large knobs that can be turned to reduce usage:

1. Reduce index coverage (exclude repositories or targets or branches) or index less frequently
2. Retain fewer indexes as they age (e.g., do not keep indexes on the non-head of branches)

The following factors also play a contributing role to the scale of data on a particular instance:

- How *large* your repositories (or indexing targets) are
- How *frequently* your repositories (or indexing targets) receive new commits
- How *widely* your data retention policies retain individual SCIP indexes (one branch vs all branches)
- How *deeply* your data retention policies retain individual SCIP indexes (the latest commit vs all commits on a branch)
- How many repositories (or indexing targets) have received SCIP uploads (are there deprecated repos being indexed that will never be relevant to a user?)

---
---
---

TODO: How best to use auto-index and CI/CD based indexing together? I would like to use CI/CD in the first instance to index (especially with Kotlin/Java where the indexing process requires a full build, which is sometimes hard outside our CI/CD infra), but use auto-indexing as fallback for repos that either don’t have CI/CD indexes, or where CI/CD has failed (i.e. the build itself is fine, but CI/CD infra had a moment and lost the job/index before upload)

TODO: Add example configuration for sg/sg, with detailed description of the chosen config, and why certain value where picked
