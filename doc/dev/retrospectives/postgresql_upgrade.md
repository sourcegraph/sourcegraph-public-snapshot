# Tomás’ notes on the PostgreSQL upgrade

## Action Items / Learnings

1. Master must always be in a releasable state. Don’t create breaking changes that block the next release.
1. Estimate and plan work before scheduling it for a release.
1. Prefer to share context using GitHub Issues instead of Slack.
1. Warning signs to be careful with adopted issues: missed estimates, being passed around, and fixed deadline.

## Background

At the time we decided to upgrade Postgres, these were the reasons for pursuing that work:

1. To use UPSERT in upcoming repo-updater work introduced in Postgres 9.5. We were running two versions: 9.6 in development, server image and CI; and 9.4 in deploy-sourcegraph-xxx.
1. More importantly, to ensure continued support and security patches. 9.x is quite old and won't be officially supported anymore soon.

## Things that didn't go well

Working with the principle of making Sourcegraph admin's lives as easy and pain-free as possible, we decided to automate the whole upgrade procedure for server image and Kubernetes installations. In retrospect, it's now clear that undertaking this work was too risky in the time-frame we had available for the 3.0 release. I took over this from Keegan as a release blocker for 3.0. We grossly underestimated the effort involved to complete this. It would never have been possible to do it in time for 3.0

> Nick says: Beyang flagged this as work we shouldn't try to squeeze in, but we tried to anyway. I think trying to squeeze this in was the avoidable root mistake.
> Beyang says: More background: Nick asked me to look into this before the holidays. I said I would timebox to one day. After looking into it for an afternoon, I decided it was more trouble than it was worth.

An even earlier root cause was that this didn't have a clear owner from the beginning of the iteration. It was a general proposal that was ownerless for awhile. I took it on, then Keegan, then Tomás.

After finishing the Postgres upgrade on sourcegraph.com, we were biased to reuse the solution we had developed for Kubernetes in the server image to save time. In retrospect, attempting to do that cost us more time because the server image required a lot of upfront work and changes to permit re-use of what we had built for the Kubernetes environment.

Particularly, unlike the Debian based image we run in Kubernetes, the server Alpine image doesn't facilitate multiple versions of Postgres to be installed side by side, which is required for the upgrade. Additionally, we were at the time building all our Docker images in sourcegraph/sourcegraph with godocockerize which was limited in its ability to let us shape the resulting container to replicate the upgrade functionality of sourcegraph/postgres-11.

In the end, we found another solution for the server image upgrade which we could have saved time with earlier had we not underestimated the effort involved in changing the server image to allow us to reuse the work done for Kubernetes in sourcegraph/postgres-11

We proceeded to upgrade to the latest stable version (11.1) without assessing the compatibility of that version with the environments of our customers that run external databases.

We rushed the first time around for the beta release on Postgres because  because we thought at the time that it was good for new users to already have 11.1 running. In retrospect, this was not really necessary and it caused more problems than it solved.

After releasing 3.0 beta with the Postgres upgrade half done, we failed to clearly communicate this state of affairs to team members. Kingy, for instance, didn't learn in time about it and proceeded to ask some existing server image customers to upgrade to the 3.0.beta which didn't work since the auto-upgrade procedure wasn't in place at that time yet.

Others, like Francis, didn't want to throw away their local Postgres data, and since they were using the server image, they had to wait until the auto upgrade was available. This blocked them from properly testing the beta locally with their data.

After releasing 3.0 beta, we decided to change my focus back to repo-updater work and to hand-over the remaining Postgres upgrade work to Keegan. The context switch costs involved further delayed the done date of the upgrade.

At the time it was planned to tag the 3.0.0 release, it was clear that we weren't going to make it in time again. Under the release pressure, we didn't consider everything to be considered in deciding to abort the upgrade entirely and revert back to 9.6. The rationale for reverting was that this upgrade wasn't bringing any immediate value to our users and it blocking the the whole release was not worth it.

Additionally, at the last minute, we learned that customers running with external Postgres instances have limitations on which version of Postgres they can run, and some couldn't run 11.1 in their environments. We also realised, through this, that we hadn't considered this type of deployments in the upgrade.

So reversing the version back to 9.6 was at that time the best decision we thought of. Unfortunately, we didn't consider that existing 3.0.beta customers were already running 11.1. This turned out to be a big problem as asking some of those high potential customers to nuke their data would likely have too much of a negative impact on their perception of us. As it turned out after a day of work, downgrading Postgres isn't officially supported and there wasn't a clear way to do it, so we decided to reverse the reversal decision.

In both the 9.6 reversal decision and the follow-up decision to reverse that previous decision, we weren't as aligned as we could have been and that created unnecessary tension.

I personally had to consistently over-work to move things forward. This was a combined result of poor time estimation and all of the above.

## Things that went well

Through working together on this, Keegan and I realised that we really like to work with each other. That resulted in deciding to join forces going forward by forming a team.

We got rid of godockerize which was an open issue of tech debt we wanted to take care of.

We (re-)introduced Postmortems as a practice to learn from operational incidents because of the first two failed attempts at upgrading Postgres in production.

Partially motivated by these events, we (re)-introduced retrospectives to continuously learn as a team about what works and what doesn't, and thus, helping us replicate our successes and avoiding to repeat our mistakes.

Despite the extended cost, we ended up succeeding in building a hands-free, pain-free Postgres upgrade path for server image and Kubernetes deployments. Admins running this sort of deployment won't have to learn and deal with the intricacies of upgrading Postgres.

## Where we got lucky

We realised that our Kubernetes Postgres deployment is using an unsafe roll-out strategy where more than one Postgres master can be up at once. Despite not having any problems with this historically, it's a risk we shouldn't have, and thus, we were lucky to notice it through the course of this work.

Through a combination of luck and the learnings with the above experiences, we found a way to auto-upgrade the server image without requiring all of the remaining Docker image work we previously thought was needed. This ended up not delaying the release of 3.0.0 as much as it would have otherwise.
