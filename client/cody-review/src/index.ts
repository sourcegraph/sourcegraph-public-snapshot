import Octokit from '@octokit/rest'

import { GraphQLResult, requestGraphQLCommon, gql } from '@sourcegraph/http-client'

import { DEFAULT_APP_SETTINGS, ENVIRONMENT_CONFIG } from './constants'
import { createCodebaseContext } from './context'
import { Message, SpeakerType } from './graphql-operations'

const { SOURCEGRAPH_SERVER_ENDPOINT } = ENVIRONMENT_CONFIG

const octokit = new Octokit()

const requestGraphQL = <TResult, TVariables = object>(
    request: string,
    variables?: TVariables
): Promise<GraphQLResult<TResult>> =>
    requestGraphQLCommon<TResult, TVariables>({
        request,
        variables,
        headers: {
            Accept: 'application/json',
            'Content-Type': 'application/json',
        },
        baseUrl: SOURCEGRAPH_SERVER_ENDPOINT,
    }).toPromise()

export const CODY_QUERY = gql`
    query CodyReview($messages: [Message!]!) {
        completions(input: { messages: $messages, temperature: 0.2, maxTokensToSample: 1000, topK: -1, topP: -1 })
    }
`

async function review(): Promise<void> {
    // Create a context for the codebase using the default app settings
    const codebaseContext = await createCodebaseContext(DEFAULT_APP_SETTINGS.codebase, DEFAULT_APP_SETTINGS.contextType)

    const { data: diff } = await octokit.pulls.get({
        owner: 'sourcegraph',
        repo: 'sourcegraph',
        pull_number: 50724,
        mediaType: {
            format: 'diff',
        },
    })

    const context = await codebaseContext.getContextMessages(diff, { numCodeResults: 5, numTextResults: 5 })

    const prompt =
        'You are a code review assistant.\n' +
        'You will now review the code changes in a pull request.\n' +
        'Please review the changes and generate suitable comments. You should format the comments in the following structure: [file name]"[relevant line number]:[relevant comment]. You do not need to repeat the code.\n' +
        'Note that you only see a small part of the diff, but you can use the context to understand the changes.\n' +
        'If your suggestion is related to changes that would typically be enforced with a lint rule, ignore it.' +
        'Here is the relevant context:\n' +
        context +
        'Here is a partial diff to review:\n' +
        diff

    const messages: Message[] = [
        {
            speaker: SpeakerType.HUMAN,
            text: prompt,
        },
        {
            speaker: SpeakerType.ASSISTANT,
            text: '',
        },
    ]

    const summary = await requestGraphQL<{ completions: string }>(CODY_QUERY, { messages })
    console.log(summary)
}

// eslint-disable-next-line @typescript-eslint/no-floating-promises
review()
