# Retrospectives

We value continuous learning and improvement.

Retrospectives are a chance for teams to reflect on past experiences, celebrate what went well, and identify areas of improvement for the future.

Significant projects and releases are great learning opportunities and should always have retrospectives. Anyone involved in a project or release can lead a retrospective.

## Completed retrospectives

<!--
Add links to completed retrospective docs here. These are publicly visible, so make sure they don't include anything sensitive.
-->

- [Sourcegraph 3.0 beta retrospective](3_0_beta.md)

## How to facilitate a retrospective

If you are facilitating a retrospective for a project or release, follow these steps.

### Set the context

Why is this retrospective being held? What will it cover? Retrospectives, similarly to any other meeting, are more effective if the participants align their expectations before getting into it.

Be sure that everyone understands the answers to the above questions before proceeding.

### Collect feedback

Identify everyone who should participate in the retrospective: this is the retrospective group. This should include everyone who participated in the project or release.

Create a [Google Form](https://docs.google.com/forms) to collect written feedback. The form should be a single input with the following prompt:
```
What feedback or thoughts do you have about <retrospective topic>? Here are some things you might want to consider:
- <Link to previous retrospective document for this topic>
- What went well? What did you like?
- What didn't go well? What didn't you like?
- Did you learn something?
```

Allocate enough time for teammates to thoughtfully respond to the survey (aim for 100% participation), but don't let the process take an unbounded amount of time. It is useful to complete the retrospective process expediently so that information is fresh in people's minds and so we can apply learnings sooner. 3 business days is a good default.

If the deadline is approaching, privately message anyone who has not submitted feedback.

> Hi $NAME! I noticed that you haven't filled out the $RETROSPECTIVE_TOPIC survey yet and was wondering if you are planning to participate? If so, can you please fill out this survey in the next 24 hours: $LINK. If not, I would appreciate any feedback about why you are not interested in participating so we can adapt our retrospective process to be more inclusive in the future.

### Share feedback

After the feedback period is over, aggregate all of the raw feedback into a single document and share it with the group so everyone can have at least one working day before the retrospective meeting to review the results.

### Retrospect, together

Schedule a meeting to discuss the submitted feedback, following this time-boxed approach.

#### Begin

Begin by asking one of the participants to read the [The Prime Directive](http://retrospectivewiki.org/index.php?title=The_Prime_Directive) out loud.

> "Regardless of what we discover, we understand and truly believe that everyone did the best job they could, given what they knew at the time, their skills and abilities, the resources available, and the situation at hand."
> --Norm Kerth, Project Retrospectives: A Handbook for Team Review

This is to assure that a retrospective has the right tone to make it a positive and result oriented event. It makes a retrospective become an effective team gathering to learn and find solutions to improve the way of working together.

#### Group (10-15 min)

The goal is to read through each raw feedback item and group related feedback items into topics.

Have the author of each feedback item read it out loud and answer any questions. Then ask the author to propose a new or existing topic for this feedback item and discuss that choice with the group.

#### Vote (5 min)

After all feedback items have been organized into topics, ask each participant vote on the topics that they deem most important. Make sure that participants understand the rules for voting:

- Each person has five votes (+1s).
- Each person can vote for a topic more than once (e.g. +2, +3).
- Each person must use all of their votes.

#### Sort (1 min)

After voting is complete, sort the topics based on the number of votes each received. This sorted list represents the collective prioritization of what to discuss.

#### Discuss (45-90 min)

The discussion is the main part of any retrospective and where real value is extracted.

Announce the time-box for this phase. Optionally, choose to time-box the discussion of each individual topic (to 5 or 10 minutes). The advantage of this approach is that it tends to keep the conversation on topic and moving at a faster pace. Don't over-do it though. There's a fine balance between keeping things on time and missing important discussion. Use your judgement.

Ensure you ask everyone what their thoughts on each topic are. Everyone is different; some people are loud and very outspoken. Some are quiet and observe more. It's important to capture everyone's opinions.

For each topic, if necessary, ensure there's a written down actionable commitment with an agreed upon owner who will pursue its resolution. Additionally, capture important discussion points for each topic in the shared document.

#### Meta-retro (5 min)

At the end of your retrospective, it’s important to take five minutes to discuss how it went. Be open and encouraging of feedback from the participants, and in the same way as the topics that were discussed, look at ways to improve the retrospective the next time around.

### Publish the results

Remove any sensitive information from the retrospective document and commit it to this directory.
