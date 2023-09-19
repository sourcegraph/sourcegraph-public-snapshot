import React, { useCallback, useState } from 'react'

import { useApolloClient } from '@apollo/client'

import { LazyQueryInput } from '@sourcegraph/branded'
import { Timestamp } from '@sourcegraph/branded/src/components/Timestamp'
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

import { preamble } from './prompt/preamble'
import { buildCommitPrompt, CommitPromptInput } from './prompt/prompt'

import styles from './DigestList.module.scss'

interface DigestListProps {
    authenticatedUser: AuthenticatedUser | null
}

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
            author {
                date
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

        // Flatten results and limit to 10 to avoid sending too many queries to Cody
        const flattenedResults = searchResults.flat().slice(0, 10)

        // Sort results by date
        const sortedResults = flattenedResults.sort(
            (a, b) => new Date(b.result.commit.author.date).getTime() - new Date(a.result.commit.author.date).getTime()
        )

        // Group and deduplicate results by repo
        const filteredGroupedResults: {
            [repo: string]: ChangelogChange[]
        } = {}

        for (const { focus, result } of sortedResults) {
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

        const changelogSummaries: ChangelogSummary = {}

        setSearchLoading(false)
        setCodyLoading(true)

        await Promise.all(
            Object.entries(filteredGroupedResults).map(async ([repo, changes]) => {
                const summaries = await Promise.all(
                    changes.map(async change => {
                        const promptInput = formatPrompt(change.result, change.focus)
                        const prompt = buildCommitPrompt({
                            input: promptInput,
                            granularity,
                        })

                        const { data } = await client.query<SummarizeTextResult, SummarizeTextVariables>({
                            query: getDocumentNode(CODY_QUERY),
                            variables: {
                                messages: [
                                    ...preamble,
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
    }, [client, granularity, queries])

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
                                                <small className="ml-2 text-muted">
                                                    <Timestamp noAbout={true} date={change.commit.author.date} />
                                                </small>
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

export const DigestList: React.FunctionComponent<React.PropsWithChildren<DigestListProps>> = () => (
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
                <ExampleChangelog
                    name="Changes for Sourcegraph app"
                    type="Commit-by-commit"
                    granularity="Detailed"
                    queries={[
                        {
                            query: 'context:global repo:^github.com/sourcegraph/sourcegraph$ type:diff message:^app: patternType:regexp after:"2 months ago"',
                            focus: 'commits and diffs',
                        },
                        {
                            query: 'context:global repo:^github.com/sourcegraph/sourcegraph$ type:diff file:client/web/src/enterprise/app/. patternType:regexp after:"2 months ago"',
                            focus: 'commits and diffs',
                        },
                        {
                            query: 'context:global repo:^github.com/sourcegraph/(about|handbook)$ type:diff \\bapp\\b patternType:regexp after:"2 months ago"',
                            focus: 'diffs',
                        },
                    ]}
                />
            </div>
        </div>
    </>
)
