import { Message, SpeakerType } from '@sourcegraph/shared/src/graphql-operations'

import { CommitPromptInput, buildCommitPrompt } from './prompt'

interface Preamble {
    input: CommitPromptInput
    output: string
}

const markdownPreamble1: Preamble = {
    input: {
        input: {
            heading: 'Update delivering-impact-reviews.md (#6630)',
            description: null,
            diff: 'content/departments/people-talent/people-ops/process/teammate-sentiment/impact-reviews/delivering-impact-reviews.md content/departments/people-talent/people-ops/process/teammate-sentiment/impact-reviews/delivering-impact-reviews.md\n@@ -4,3 +4,3 @@ \n \n-Impact reviews will be delivered synchronously in a 1:1 between the Manager and their direct report. Each Manager is responsible for scheduling a 30 - 60 minute (recommended) meeting with each Teammate to deliver their review packet, along with any corresponding promotion or compensation increases. All conversations must take place no later than **October 14 at the latest.**\n+Impact reviews will be delivered synchronously in a 1:1 between the Manager and their direct report. Each Manager is responsible for scheduling a 30 - 60 minute (recommended) meeting with each Teammate to deliver their review packet, along with any corresponding promotion or compensation increases. All conversations must take place no later than \\*_April 26, 2023 for H1 FY24 Impact Review Cycle_\n \n',
        },
        granularity: 'Overview',
    },
    output: `
- Updated the impact review process
- Old deadline: October 14
- New deadline: April 26, 2023 (H1 FY24 Impact Review Cycle)`,
}

const markdownPreamble2: Preamble = {
    input: {
        input: {
            heading: 'updates customer information (#6625)',
            description: 'updated private with more examples of customer information',
            diff: 'content/company-info-and-process/policies/data-sharing.md content/company-info-and-process/policies/data-sharing.md\n@@ -53,3 +53,3 @@ Below you can find a matrix to help you make informed decisions about what data\n    </td>\n-   <td>Customer private source code\n+   <td>Customer private source code snippets (for support purposes)\n    </td>\n@@ -63,3 +63,3 @@ Below you can find a matrix to help you make informed decisions about what data\n    </td>\n-   <td>private repository names, legal contracts, company financials, incident reports for security issues \n+   <td>Customer roadmaps, customer number of codebases, customer challenges, private repository names, legal contracts, company financials, incident reports for security issues, private repository names, legal contracts, company financials, incident reports for security issues \n    </td>\n',
        },
        granularity: 'Overview',
    },
    output: `
- Updated customer information in the data-sharing policy
- Added more examples of private customer information
- Examples include:
  - Customer roadmaps
  - Number of customer codebases
  - Customer challenges
  - Private repository names (repeated)
  - Legal contracts (repeated)
  - Company financials (repeated)
  - Incident reports for security issues (repeated)
  - Customer private source code snippets (for support purposes)
  - This change updated the customer information policy.`,
}

const instructions = `
You are an expert at understanding changes and building a clear, concise and readable summary.

You will follow these rules strictly:
1. Use all of the **relevant** information available to build your summary.
2. If the user specifies that the summary should have an "Overview" granularity, then you should only include the most important changes.
3. If the user specifies that the summary should have a "Detailed" granularity, then you should include all of the changes.
4. Format your summary in a bullet-point list. Do not use any other formatting.
5. Do not mention details like specific files changed or commit hashes.
6. Note that the diff is only a small preview of the most relevant parts of the change. Avoid assuming too much.
7. Don't try to provide your own introduction or conclusion. The summary should be a standalone list of changes. For example, don't prefix your response with "Here is a summary of the changes" or any other similar introduction.
`

export const preamble: Message[] = [
    {
        speaker: SpeakerType.HUMAN,
        text: instructions,
    },
    {
        speaker: SpeakerType.ASSISTANT,
        text: 'Understood. I am an expert at understanding changes and building a clear, concise and readable summary. I will try to help you understand a change, and ensure I follow my strict rules when doing so.',
    },
    {
        speaker: SpeakerType.HUMAN,
        text: buildCommitPrompt(markdownPreamble1.input),
    },
    {
        speaker: SpeakerType.ASSISTANT,
        text: markdownPreamble1.output,
    },
    {
        speaker: SpeakerType.HUMAN,
        text: buildCommitPrompt(markdownPreamble2.input),
    },
    {
        speaker: SpeakerType.ASSISTANT,
        text: markdownPreamble2.output,
    },
]
