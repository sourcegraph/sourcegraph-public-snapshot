# Retrospectives

We value continuous learning and improvement.

Retrospectives are a chance for teams to reflect on past experiences, celebrate what went well, and identify areas of improvement for the future.

The [release captain](../releases.md#release-captain) is responsible for facilitating an engineering wide retrospective for each monthly release, but additional retrospectives may be organized by anyone for any reason (e.g. team retrospective, ops incident, customer incident).

## How to facilitate a retrospective

If you are facilitating a retrospective for a project or release, follow these steps.

### Collect feedback

Create a new Google Doc with an appropriate title (e.g. "3.2 retrospective") that is editable by everyone who is participating in the retrospective. Paste in the following agenda:

```
Purpose: [SHORT DESCRIPTION OF THE SCOPE AND PURPOSE OF THIS RETROSPECTIVE]
Retrospective meeting: [DATE AND TIME]
Please submit feedback 24 hours before the retrospective meeting so that everyone can pre-read the feedback.

"Regardless of what we discover, we understand and truly believe that everyone did the best job they could, given what they knew at the time, their skills and abilities, the resources available, and the situation at hand."
--Norm Kerth, Project Retrospectives: A Handbook for Team Review

What feedback or thoughts would you like to share with the team? Here are some things you might want to consider:
- [LINK TO RELEVANT PREVIOUS RETROSPECTIVE]
- What went well? What did you like?
- What didn't go well? What didn't you like?
- Did you learn something?
- What do you wish you had done differently?

Add your feedback by editing this document directly.
```

Send a Slack message and a calendar invite for the retrospective meeting that includes a link to this document.

If it is getting close to or past the 24 hour feedback deadline, privately remind anyone who has not submitted feedback.

### Retrospective meeting

You responsible for the agenda and flow of the retrospective meeting.

Share your screen which has the retrospective document open and a [timer](https://www.google.com/search?q=timer) to keep the meeting on track.

#### Begin

Begin by asking one of the participants to read the [The Prime Directive](http://retrospectivewiki.org/index.php?title=The_Prime_Directive) out loud.

> "Regardless of what we discover, we understand and truly believe that everyone did the best job they could, given what they knew at the time, their skills and abilities, the resources available, and the situation at hand."
> --Norm Kerth, Project Retrospectives: A Handbook for Team Review

This is to assure that a retrospective has the right tone to make it a positive and result oriented event. It makes a retrospective become an effective team gathering to learn and find solutions to improve the way of working together.

#### Review (5 min)

Review the action items from the previous retrospective and discuss the current state. Is it complete? If not, why? Are there any followup actions?

#### Group (15 min)

Have the author of each feedback item read it out loud and answer any questions or points of clarification (this is not time for discussion).

#### Vote (5 min)

Ask each participant vote on the items that they deem most important. Make sure that participants understand the rules for voting:

- Each person has ten votes (+1s).
- Each person can vote for a topic more than once (e.g. +2, +3).
- Each person must use all of their votes.

#### Sort (1 min)

After voting is complete, sort the items based on the number of votes each received. This sorted list represents the collective prioritization of what to discuss.

#### Discuss (~60 min)

The discussion is the main part of any retrospective and where real value is extracted.

This phase lasts for the rest of the scheduled meeting (minus 5 minutes for the meta retro).

Based on the vote distribution, choose which topics are going to be discussed and the time limit for each topic. 10 minutes is a good default. Enforce the time limit! You can always return to a topic at the end if there is extra time.

Ensure you ask everyone what their thoughts on each topic are. Everyone is different; some people are loud and very outspoken. Some are quiet and observe more. It's important to capture everyone's opinions.

For each topic, if necessary, ensure there's a written down actionable commitment with an agreed upon owner who will pursue its resolution. Additionally, capture important discussion points for each topic in the shared document.

### Publish the results

Remove any sensitive information from the retrospective document and commit it to this directory.

To maintain list formatting, download the Google Doc as plaintext instead of attempting to copy/paste: File > Download As > Plain Text (.txt).

## Completed retrospectives

<!--
Add links to completed retrospective docs here. These are publicly visible, so make sure they don't include anything sensitive.
-->

- [3.0 beta](3_0_beta.md)
- [3.0](3_0.md)
- [Tomás’ notes on the PostgreSQL upgrade](postgresql_upgrade.md)
- [3.2](3_2.md)
- [Customer license expiration](customer_license_expiration.md)
- [3.3](3_3.md)
- [3.4](3_4.md)
- [3.5](3_5.md)
- [3.6](3_6.md)
- [3.7](3_7.md)
- [3.8](3_8.md)
- [3.9](3_9.md)
