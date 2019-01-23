# Better in-app guidance for (1) admins to set up code intelligence and (2) users to enable it

A plan to provide a more guided path in-app to enabling and getting value from code intelligence.

## Background

Sourcegraph has never had much onboarding, but prior to Sourcegraph 3.0.0, getting value from code intelligence was somewhat more straightforward:
* Most language support "just worked" automatically, without the admin doing anything.
* We used to have a **Site-admin > Code intelligence** page, which at least provided the admin with an obvious place to go if something wasn't enabled or working.

Today, there basically is no way to discover code intelligence unless the admin goes searching for it in the docs (or happens across the relevant extension).

## Plan

1) **First** (3.0), we add a post-install site alert for admins only that links them to http://sourcegraph.example.com/extensions?query=category%3A%22Programming+languages%22.

This would appear at the top of every page (after the admin adds repos and the "Add repositories" alert goes away) and it would say something like "Enable Sourcegraph extensions to get code intelligence for your team"

2) **Second** (3.1), we add a new type of toast (using https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/blob/web/src/marketing/Toast.tsx) for all users that appears programmatically when a user visits a new file extension (.go, .ts, etc.) that links them to the relevant extension (or basic codeintel) if they don't have it enabled. If the user is an admin, the toast would be enhanced with a link to docs on language server deployment (though no obvious page exists for this in the admin docs today, see below). 

These would be dismissable by language. This work would provide the foundation for future extension recommendations.

3) **Third** (3.0), we make each language extension's README follow the same format:

- Table of contents
- Image showing a hover tooltip
- Who this README is for (admins)
- Instructions for running in Docker and Kubernetes
- Then information specific to the language server (e.g. private dependencies)

This is necessary because of how busy/complicated the extension READMEs have already become. There are so many potential paths on those pages (e.g.: admin setup vs. user extension enablement; sourcegraph.com vs. self-hosted; 3.0 vs. 2.x; single Docker image vs. cluster, etc.). Pulling the admin setup portion out of those READMEs, and leaving them to focus on how end users (i.e., non-admins) enable the extension, would simplify and clarify them. They would, of course, link out to the docs page.

### Test plan

TBD. We could add an e2e test to ensure the site alert and toast appear for admins and users at the right moments.

### Release plan

These would be low-risk additions. They would be released to all users as soon as they were available.

### Success metrics

Interaction rates on Sourcegraph.com (i.e., users who click through the toast to enable code intellgience for a given language),and usage of code intelligence on self-hosted instances. 

Also, where difficult to measure, direct feedback from existing, friendly customers.

### Company goals

This would directly drive usage of our most popular feature, code intelligence, and would provide a much faster (and more obvious) path to value for new users. This all serves to make it much more likely that an installer would share with their teammates and grow the instance from 1 to 20 users.

## Rationale

There are a lot of ways we could make this easier for admins and end-users — these are just a proposal for a package of improvements, but are open to suggestions.

Other ideas include different UIs for recommendations (e.g. using something other than site alerts and toasts), using emails, adding different ways to split up extension READMEs to make it easier to follow as an admin vs. as a user, etc. 

## Checklist 

TBD, pending team approval of the proposals.

## Done date

Documentation and site alert can be implemented by the 3.0 launch.

Recommendations (toasts) would be released in 3.1 (early  March).

## Retrospective

[This section is completed after the project is completely done (i.e. the checklist is complete).]

### Actual checklist

[The checklist that was actually completed (i.e. paste the final checklist from the issue). Explain any differences from the original checklist in the plan.]

### Actual done date

[The date that the project was actually finished. Explain why this is earlier or later than originally planned or explain why the project was not completed.]
