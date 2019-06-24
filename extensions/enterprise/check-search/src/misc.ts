import * as sourcegraph from 'sourcegraph'

export const OTHER_CODE_ACTIONS: sourcegraph.CodeAction[] = [
    {
        title: 'Open tsconfig.json',
        command: {
            title: '',
            command: 'open',
            arguments: ['http://localhost:3080/github.com/sourcegraph/sourcegraph/-/blob/web/tsconfig.json'],
        },
    },
    {
        title: 'Start discussion thread',
        command: { title: '', command: 'TODO!(sqs)' },
    },
    {
        title: 'Message code owner: @tsenart',
        command: { title: '', command: 'TODO!(sqs)' },
    },
]

export const REPO_INCLUDE = '(sourcegraph2|codeintellify|about|react-loading-spinner)'

export const MAX_RESULTS = 10
