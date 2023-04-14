import React, { useCallback, useState } from 'react'

import { useApolloClient } from '@apollo/client'

import { LazyQueryInput } from '@sourcegraph/branded'
import { CHARS_PER_TOKEN, MAX_RECIPE_INPUT_TOKENS } from '@sourcegraph/cody-shared/src/prompt/constants'
import { truncateText } from '@sourcegraph/cody-shared/src/prompt/truncation'
import { renderMarkdown } from '@sourcegraph/common'
import { getDocumentNode, gql } from '@sourcegraph/http-client'
import { Container, Link, H2, H3, Button, Markdown, LoadingSpinner, ErrorAlert, Label } from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../../auth'
import {
    SearchCommitSearchResult,
    SearchForChangesResult,
    SearchForChangesVariables,
    SearchPatternType,
    SpeakerType,
    SummarizeTextResult,
    SummarizeTextVariables,
} from '../../graphql-operations'

import { CodeMonitoringPageProps } from './CodeMonitoringPage'

import styles from './ChangelogAnywhereList.module.scss'

interface ChangelogAnywhereListProps
    extends Required<Pick<CodeMonitoringPageProps, 'fetchUserCodeMonitors' | 'toggleCodeMonitorEnabled'>> {
    authenticatedUser: AuthenticatedUser | null
}

const markdownPreamble1 = {
    input: {
        question: 'What has changed in our handbook this week?',
        granularity: 'Overview',
        heading: 'Update delivering-impact-reviews.md (#6630)',
        description: null,
        diff: 'content/departments/people-talent/people-ops/process/teammate-sentiment/impact-reviews/delivering-impact-reviews.md content/departments/people-talent/people-ops/process/teammate-sentiment/impact-reviews/delivering-impact-reviews.md\n@@ -4,3 +4,3 @@ \n \n-Impact reviews will be delivered synchronously in a 1:1 between the Manager and their direct report. Each Manager is responsible for scheduling a 30 - 60 minute (recommended) meeting with each Teammate to deliver their review packet, along with any corresponding promotion or compensation increases. All conversations must take place no later than **October 14 at the latest.**\n+Impact reviews will be delivered synchronously in a 1:1 between the Manager and their direct report. Each Manager is responsible for scheduling a 30 - 60 minute (recommended) meeting with each Teammate to deliver their review packet, along with any corresponding promotion or compensation increases. All conversations must take place no later than \\*_April 26, 2023 for H1 FY24 Impact Review Cycle_\n \n',
    },
    output: `
- Updated the impact review process
  - Old deadline: October 14
  - New deadline: April 26, 2023 (H1 FY24 Impact Review Cycle)
    `,
}

const markdownPreamble2 = {
    input: {
        question: 'What has changed in our handbook this week?',
        granularity: 'Overview',
        heading: 'updates customer information (#6625)',
        description: 'updated private with more examples of customer information',
        diff: 'content/company-info-and-process/policies/data-sharing.md content/company-info-and-process/policies/data-sharing.md\n@@ -53,3 +53,3 @@ Below you can find a matrix to help you make informed decisions about what data\n    </td>\n-   <td>Customer private source code\n+   <td>Customer private source code snippets (for support purposes)\n    </td>\n@@ -63,3 +63,3 @@ Below you can find a matrix to help you make informed decisions about what data\n    </td>\n-   <td>private repository names, legal contracts, company financials, incident reports for security issues \n+   <td>Customer roadmaps, customer number of codebases, customer challenges, private repository names, legal contracts, company financials, incident reports for security issues, private repository names, legal contracts, company financials, incident reports for security issues \n    </td>\n',
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
  - This change updated the customer information policy.
    `,
}

const codyRules = `
1. Use all of the **relevant** information available to build your summary.
2. If the user specifies that the summary should have an "Overview" granularity, then you should only include the most important changes.
3. If the user specifies that the summary should have a "Detailed" granularity, then you should include all of the changes.
4. Format your summary in a bullet-point list. Do not use any other formatting.
5. Do not mention details like specific files changed or commit hashes.
6. Note that the diff is only a small preview of the most relevant parts of the change. Avoid assuming too much.
7. Don't try to provide your own introduction or conclusion. The summary should be a standalone list of changes. For example, don't prefix your response with "Here is a summary of the changes" or any other similar introduction.
8. Omit any information that is irrelevant or unimportant to the user's goal.
`

const humanCommitPreamble = `
I want you to summarize a change for me, here's an example of a previous conversation we had, so you can understand what to do:

Human:

${JSON.stringify(markdownPreamble1.input)}

Generate a high-level summary of this change in a readable, plaintext, bullet-point list.

Additional information to help you build your summary:
- The user is trying to answer the question: ${markdownPreamble1.input.question}
- The summary should have the following granularity: ${markdownPreamble1.input.granularity}

Follow these rules strictly:
${codyRules}

Assistant:

${markdownPreamble1.output}

Human:

${JSON.stringify(markdownPreamble2.input)}

Generate a high-level summary of this change in a readable, plaintext, bullet-point list.

Additional information to help you build your summary:
- The user is trying to answer the question: ${markdownPreamble2.input.question}
- The summary should have the following granularity: ${markdownPreamble2.input.granularity}

Follow these rules strictly:
${codyRules}
`

const assistantCommitPreamble = `
Assistant:

${markdownPreamble2.output}
`

interface CommitPromptInput {
    input: {
        heading: string
        description: string | null
        diff: string | null
    }
    goal: string
}

export const getCommitPrompt = (input: CommitPromptInput['input'], question: string, granularity: string): string => `
Human:

${JSON.stringify(input)}

Generate a summary of this change in a readable, plaintext, bullet-point list.

Additional information to help you build your summary:
- The user is trying to answer the question: ${question}
- The summary should have the following granularity: ${granularity}

Follow these rules strictly:
${codyRules}

Assistant:

-`

const SEARCH_QUERY = gql`
    query SearchForChanges($query: String!) {
        search(query: $query) {
            results {
                results {
                    ... on CommitSearchResult {
                        ...SearchCommitSearchResult
                    }
                }
            }
        }
    }

    fragment SearchCommitSearchResult on CommitSearchResult {
        url
        diffPreview {
            value
        }
        commit {
            id
            repository {
                name
            }
            subject
            body
        }
    }
`

/**
 * What to focus on when generating the changelog.
 * - commits: only use the commit message subject and body
 * - diffs: only use the diff (currently just `diffPreview`)
 * - commits and diffs: use both the commit message and the diff
 */
type ChangelogFocus = 'commits' | 'diffs' | 'commits and diffs'

function formatPrompt(result: SearchCommitSearchResult, focus: ChangelogFocus): CommitPromptInput['input'] {
    let promptHeading: CommitPromptInput['input']['heading'] = ''
    let promptDescription: CommitPromptInput['input']['description'] = null
    let promptDiff: CommitPromptInput['input']['diff'] = null
    let maxPromptLength = MAX_RECIPE_INPUT_TOKENS * CHARS_PER_TOKEN

    if (focus === 'commits' || focus === 'commits and diffs') {
        promptHeading = truncateText(result.commit.subject, maxPromptLength)
        maxPromptLength = maxPromptLength - promptHeading.length

        promptDescription = result.commit.body ? truncateText(result.commit.body, maxPromptLength) : null
        maxPromptLength = maxPromptLength - (promptDescription ? promptDescription.length : 0)
    }

    if (focus === 'diffs' || focus === 'commits and diffs') {
        promptDiff = result.diffPreview?.value ? truncateText(result.diffPreview?.value, maxPromptLength) : null
    }

    return {
        heading: promptHeading,
        description: promptDescription,
        diff: promptDiff,
    }
}

function removeEmptyBulletPoints(text: string): string {
    const lines = text.split('\n')
    const filteredLines = lines.filter(line => line.trim() !== '-' && line.trim() !== '')
    return filteredLines.join('\n')
}

/**
 * Some basic filtering to get around some difficult to fix flaws with Cody's response.
 * TODO: Figure out a way to omit these from Cody's output.
 */
const filterResponse = (response: string): string => {
    const bulletPointRegex = /[•›-]/
    const firstBulletPointIndex = response.search(bulletPointRegex)
    const withoutPrefix = response.slice(firstBulletPointIndex)
    const withoutEmptyBulletPoints = removeEmptyBulletPoints(withoutPrefix)

    // Filter out any responses that are only a few characters.
    // TODO: Better way for Cody to output a "no summary" response.
    if (withoutEmptyBulletPoints.length < 10) {
        return ''
    }

    return withoutEmptyBulletPoints
}

interface ChangelogQuery {
    focus: ChangelogFocus
    query: string
}

interface ChangelogChange {
    focus: ChangelogFocus
    result: SearchCommitSearchResult
}

interface ChangelogSummary {
    [repo: string]: {
        change: SearchCommitSearchResult
        summary: string
    }[]
}

const CODY_QUERY = gql`
    query SummarizeText($messages: [Message!]!) {
        completions(input: { maxTokensToSample: 1000, temperature: 0.1, topK: -1, topP: -1, messages: $messages })
    }
`

interface ExampleChangelogProps {
    name: string
    queries: ChangelogQuery[]
    // The type of changelog to output. TODO: Other types?
    type: 'Commit-by-commit'
    // The desired amount of detail in the changelog.
    granularity: 'Overview' | 'Detailed'
}

const ExampleChangelog: React.FunctionComponent<ExampleChangelogProps> = ({ name, queries, type, granularity }) => {
    const client = useApolloClient()
    const [searchLoading, setSearchLoading] = useState(false)
    const [codyLoading, setCodyLoading] = useState(false)
    const [error, setError] = useState<Error | null>(null)
    const [changelogResults, setChangelogResults] = useState<ChangelogSummary | null>(null)

    const searchForCommits = useCallback(async () => {
        setSearchLoading(true)

        // Resolve all search queries
        const searchResults = await Promise.all(
            queries.map(async ({ query, focus }) => {
                const { data } = await client.query<SearchForChangesResult, SearchForChangesVariables>({
                    query: getDocumentNode(SEARCH_QUERY),
                    variables: {
                        query,
                    },
                })

                const results = data?.search?.results.results

                if (!results) {
                    return []
                }

                return results
                    .filter((result): result is SearchCommitSearchResult => result.__typename === 'CommitSearchResult')
                    .map(result => ({
                        focus,
                        result,
                    }))
            })
        )

        // Group and deduplicate results by repo
        const filteredGroupedResults: {
            [repo: string]: ChangelogChange[]
        } = {}

        for (const searchResult of searchResults) {
            for (const { focus, result } of searchResult) {
                const repo = result.commit.repository.name

                if (!filteredGroupedResults[repo]) {
                    filteredGroupedResults[repo] = []
                }

                const existingResult = filteredGroupedResults[repo].find(
                    existingResult => existingResult.result.commit.id === result.commit.id
                )

                if (existingResult) {
                    // If we have conflicting a focus, we want to use the more general one
                    // We probably want to revise this to support using both in the future.
                    existingResult.focus = existingResult.focus !== focus ? 'commits and diffs' : existingResult.focus
                    continue
                }

                filteredGroupedResults[repo].push({
                    focus,
                    result,
                })
            }
        }

        const changelogSummaries: ChangelogSummary = {}

        setSearchLoading(false)
        setCodyLoading(true)

        // TODO: Add limit to 10

        await Promise.all(
            Object.entries(filteredGroupedResults).map(async ([repo, changes]) => {
                const summaries = await Promise.all(
                    changes.map(async change => {
                        const promptInput = formatPrompt(change.result, change.focus)
                        const prompt = getCommitPrompt(promptInput, name, granularity)

                        const { data } = await client.query<SummarizeTextResult, SummarizeTextVariables>({
                            query: getDocumentNode(CODY_QUERY),
                            variables: {
                                messages: [
                                    {
                                        speaker: SpeakerType.HUMAN,
                                        text: humanCommitPreamble,
                                    },
                                    {
                                        speaker: SpeakerType.ASSISTANT,
                                        text: assistantCommitPreamble,
                                    },
                                    {
                                        speaker: SpeakerType.HUMAN,
                                        text: prompt,
                                    },
                                ],
                            },
                        })

                        const summary = filterResponse(data?.completions)

                        return {
                            change: change.result,
                            summary,
                        }
                    })
                )

                changelogSummaries[repo] = summaries
            })
        )

        setCodyLoading(false)
        setChangelogResults(changelogSummaries)
    }, [client, granularity, name, queries])

    return (
        <Container className="mt-2">
            <div>
                <H2>{name}</H2>
            </div>
            <div>
                <Container className="p-3 mt-3">
                    <H3>Aggregate:</H3>
                    {queries.map(({ query, focus }) => (
                        <div className="mt-1 d-flex align-items-center" key={query}>
                            <Label className="mb-0 mr-1">{focus} from:</Label>
                            <LazyQueryInput
                                className={styles.input}
                                patternType={SearchPatternType.standard}
                                caseSensitive={false}
                                isSourcegraphDotCom={window.context.sourcegraphDotComMode}
                                queryState={{
                                    query,
                                }}
                                onChange={() => null}
                                preventNewLine={false}
                                autoFocus={false}
                                editorOptions={{
                                    readOnly: true,
                                }}
                                applySuggestionsOnEnter={true}
                            />
                        </div>
                    ))}
                </Container>
                <Container className="p-3">
                    <H3>Summarize:</H3>
                    <div className="mt-1 d-flex">
                        <Label className="mb-0 mr-1">Type:</Label>
                        {type}
                    </div>
                    <div className="mt-1 d-flex">
                        <Label className="mb-0 mr-1">Granularity:</Label>
                        {granularity}
                    </div>
                </Container>
                <Container className="p-3 mb-3">
                    <H3>Notify:</H3>
                    <div className="mt-1 d-flex flex-column">
                        <ul className="mb-0">
                            <li>
                                Send <u>email to multiple recipients</u> every Monday at 9.00AM UTC.
                            </li>
                            <li>
                                Post to <u>Slack</u> every Monday at 9.00AM UTC.
                            </li>
                        </ul>
                    </div>
                </Container>

                <div className="mt-1 d-flex align-items-center justify-content-end">
                    <Button variant="secondary" onClick={() => null} className="mr-2">
                        Edit
                    </Button>
                    <Button
                        variant="primary"
                        onClick={searchForCommits}
                        disabled={changelogResults !== null || searchLoading || codyLoading}
                        className=""
                    >
                        Preview (last 10 changes)
                    </Button>
                </div>
            </div>
            <div className="d-flex flex-column align-items-center mt-2">
                {searchLoading ? (
                    <>
                        <LoadingSpinner />
                        <small>Looking for changes...</small>
                    </>
                ) : codyLoading ? (
                    <>
                        <LoadingSpinner />
                        <small>Generating changelog...</small>
                    </>
                ) : error ? (
                    <>
                        <ErrorAlert error={error} />
                    </>
                ) : changelogResults ? (
                    <Container className="w-100">
                        {Object.keys(changelogResults).map(repo => (
                            <div key={repo}>
                                <H3>{repo}</H3>
                                {changelogResults[repo]
                                    .filter(({ summary }) => Boolean(summary))
                                    .map(({ change, summary }) => (
                                        <div className={styles.change} key={change.commit.id}>
                                            <div className="d-flex align-items-center mb-1">
                                                <Label className="mr-1 mb-0">Relevant commit:</Label>
                                                <Link to={change.url}>{change.commit.subject}</Link>
                                            </div>
                                            <div>
                                                <div className="d-flex flex-column">
                                                    <Label className="mb-0">Relevant changes:</Label>
                                                    <Markdown
                                                        className={styles.output}
                                                        dangerousInnerHTML={renderMarkdown(summary)}
                                                    />
                                                </div>
                                            </div>
                                        </div>
                                    ))}
                            </div>
                        ))}
                    </Container>
                ) : (
                    <></>
                )}
            </div>
        </Container>
    )
}

export const ChangelogAnywhereList: React.FunctionComponent<
    React.PropsWithChildren<ChangelogAnywhereListProps>
> = () => (
    <>
        <div className="row mb-5">
            <div className="d-flex flex-column w-100 col">
                <div className="d-flex align-items-center justify-content-between">
                    <H3 className="mb-2">Your changelogs</H3>
                </div>
                <ExampleChangelog
                    name="What has changed with Cody?"
                    type="Commit-by-commit"
                    granularity="Overview"
                    queries={[
                        {
                            query: 'patterntype:regexp repo:^github.com/sourcegraph/sourcegraph$ type:diff after:"1 day ago" file:client/cody/.',
                            focus: 'commits and diffs',
                        },
                        {
                            query: 'patterntype:regexp repo:^github.com/sourcegraph/handbook$ type:diff after:"1 day ago" Cody',
                            focus: 'diffs',
                        },
                        {
                            query: 'patterntype:regexp repo:^github.com/sourcegraph/about$ type:diff after:"1 day ago" Cody',
                            focus: 'diffs',
                        },
                    ]}
                />
                <ExampleChangelog
                    name="What changed in the handbook in the last week?"
                    type="Commit-by-commit"
                    granularity="Overview"
                    queries={[
                        {
                            query: 'patterntype:regexp repo:^github.com/sourcegraph/handbook$ type:diff after:"1 week ago"',
                            focus: 'commits and diffs',
                        },
                    ]}
                />
                <ExampleChangelog
                    name="What has changed in our GraphQL API this week?"
                    type="Commit-by-commit"
                    granularity="Detailed"
                    queries={[
                        {
                            query: 'patterntype:regexp repo:^github.com/sourcegraph/sourcegraph$ file:..graphql$ type:diff after:"last week"',
                            focus: 'diffs',
                        },
                    ]}
                />
                <ExampleChangelog
                    name="What frontend changes have there been since yesterday?"
                    type="Commit-by-commit"
                    granularity="Detailed"
                    queries={[
                        {
                            query: 'patterntype:regexp repo:^github.com/sourcegraph/sourcegraph$ type:diff lang:TypeScript after:"yesterday"',
                            focus: 'diffs',
                        },
                    ]}
                />
            </div>
        </div>
    </>
)
