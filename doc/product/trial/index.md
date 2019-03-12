# How to run a Sourcegraph trial

Sourcegraph has been deployed inside of thousands of organizations. From our experience working with them, we have collected a set of recommendations for making a trial successful.

> Note: Trying Sourcegraph during a Hackathon? Check out our [Hackathon package](https://about.sourcegraph.com/hackathons) and let us know!

## 1. Define trial success

What? You shouldn't start with [`docker run...`](/)?

The single most important predictor of a successful trial is agreeing up-front why your organization needs Sourcegraph, and how you'll judge the results.

Trials typically last **2 to 4 weeks**, and include **25 to 500 initial users (i.e., one full, broad team or organization).** 

For measuring a trial, each company will be different. Common examples we find include:

* **Technical validation**
  When a company is replacing an existing and outdated search tool, sometimes simply validating that the deployment works, and is performant, is all that matters.

* **User engagement**
  Sourcegraph provides usage statistics to site admins, at https://sourcegraph.example.com/site-admin/usage-stats. This page will show you how many unique users used the product per day, week, and month, along with per-user metrics so you can validate the numbers. 

  Many larger customers rely on usage metrics, such as the percentage of trial users that come back to use Sourcegraph per week or day, to determine whether it was successful.

* **User surveys**
  Sourcegraph provides a built-in survey at https://sourcegraph.example.com/survey, with summary results visible to management at https://sourcegraph.example.com/site-admin/survey-results. These surveys provide bot quantitative (NPS score) and qualitative user feedback.

* **Just taking Sourcegraph away, and seeing who yells**
  While we don't always encourage this, it can be instructive! Take Sourcegraph away for a day, and see who gets upset. Developers may feel like their superpowers were suddenly taken away.

What did your company use? [Tweet at us](https://twitter.com/@srcgraph), and we'll send you some Sourcegraph swag!

Interested in learning how Sourcegraph can help your team? [Please reach out](/contact) — we love talking to prospective users, and we can share how some of the world's leading software companies use Sourcegraph internally.

## 1. Deploy Sourcegraph

Identify the correct site admin(s) — the person or team responsible for owning the technical deployment — so they can [deploy Sourcegraph internally](/site-admin/deployment).

Site admins can get direct support from Sourcegraph Engineers in our open source repository's [public issue tracker](https://github.com/sourcegraph/sourcegraph/issues).

## 1. Share with the trial team

As mentioned above, Sourcegraph trials are most successful when they start with a full team or organization (25 to 500, or more, users). If the cohort with access is smaller than that, we find users feel uncomfortable:
  (1) investing in the product and exploring features, as they don't expect it to stick around, and;
  (2) sharing links, telling others about it, posting search results, and other socializing.

If you want to run a trial with more than 100 users on your instance, just [reach out](/contact), fill us in on your goals, and we'll provide an unlimited license key for up to a month, no questions asked. 

**The bare minimum for a successful trial is providing access to 25 users.**

A typical message to the team looks like:

>Team
>
>...
>

## 1. Deploy integrations

Sourcegraph is most useful when it's at your fingertips. Our [integrations](), including our Chrome and Firefox extensions, provide code intelligence in GitHub, GitLab, Bitbucket, Phabricator, and more, and also provide a search shortcut from the browser URL bar (the "omnibox", in Chrome terminology):

[pic of Omnibox]

For short trials, however, companies often choose to only set up our omnibox integration (though it provides less functionality, it may require fewer internal approvals). See [our guide for setting up omnibox shortcuts]().

## 1. Measure success

Like step 1, this looks different for every company. If you've defined success as a 60% net promoter score (NPS), or 50%+ of your trial team using Sourcegraph every week, this is the time to measure.

## Success?

If your team found Sourcegraph to be a success, let us know. We'd love to talk about how to grow Sourcegraph ...
