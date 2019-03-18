# Towards a brighter future for repo-updater

Repo-updater's behavior is now a bit saner, and we have a better framework
to bolt things to, but endgame testing revealed an entire _category_ of
repositories it doesn't track for updates, namely, repos added directly
via "add GitHub repository" or the like.

This is a wishlist to use during iteration planning. I don't know how
long some of this will take. I can see the mud puddles, but I don't
know how deep they are.

## Task: Fix interactions between repo sources

Right now, repo-updater is doing three main things:
_ updating lists of repos from code hosts and config.json
_ doing API query lookups \* performing fetches

Only the "updating lists of repos" and "performing fetches" tasks are
deeply intertwined and don't make sense that way.

So the first obvious task is to clean this up, so that the lists
of repos being fetched are used to populate/update the repo database,
and the repo database is what gets used to determine what repos
to fetch. Decoupling those functions will make the code for both of them
a lot simpler.

## Task: Association between repos and code host configs

Repo-updater behaves poorly if you have two github.com tokens and they
can access different repositories. This logic needs to be rethought.

## Task: Interactions with gitserver cloning

Automatic cloning in gitserver can overload the clone concurrency
queue and prevent repo-updater from working. We should reduce or remove
most of the automatic cloning from gitserver, or more immediately,
make it contingent on repo-updater not doing automatic/background
fetches. In the long run, repo-updater should _still_ own those
clone operations, too, and they should have controllable interactions
with the rest of the scheduler.

## Task: Metrics for fetching

Our fetch metrics are awful. I'd like to track information which is
more useful for determining the state of the system. If the only metric
you have is staleness, the solution is always "more fetches".

Things I'd like to track:
_ Proportion of fetches which get an update.
_ Estimated staleness of new commits (when it shows up, how
long had it been there?).
_ Total number of fetches running.
_ Relative proportion of updates from background/automatic
fetches.
_ Interactions between manual and automatic fetches; how many
automatic fetches get skipped because a direct request came in?
_ Interactions between repo-updater initiated fetches, and \* automatic fetches.

The goal here is to have something that a user can look at that gives
them some insight into the state of the system. How many fetches is
repo-updater running, how long are they taking, how stale are repos,
etcetera. This would eventually lead to tuning knobs; in the short term,
probably starting with a single knob to let you adjust preferred fetch
concurrency for background updates.

### Digression about timestamps

We can't get an exact time, because there's nothing you can do to
a git repo to tell you reliably when the last commit got added,
only the datestamp of that commit, which could be years off from
when it showed up, or the time at which you checked and found
that commit and hadn't previously had it.) A good estimate of the
staleness of a commit is _probably_ "roughly half the time since the
last fetch".

## Task: Explore API overload.

Repo-updater on dogfood is making something over 40k requests to GitHub
per half hour, last I checked. This is insane. We need to figure out why
the internal caching isn't catching this, and possibly redesign that
caching. At least we fixed the HTTP cache so we're not using up our entire
rate limit in 20 minutes.

## Task: General refactoring

There's a lot of fairly duplicative code, which would be nice to clean
up in passing. Candidates are things like the RunFooSyncWorker calls
in the code host code, or the way GetRepository calls GetGitHubRepository,
GetGitLabRepository, GetBitBucketRepository, and so on.
