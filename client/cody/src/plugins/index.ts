import { Message } from '@sourcegraph/cody-shared/src/sourcegraph-api'

interface IPluginContextData {
    url: string
}

interface IPlugin {
    name: string
    modelDescription: string
    // todo: support different argument schemas
    search: (query: string) => IPluginContextData[]
}

export const PLUGINS: IPlugin[] = [
    {
        name: 'Confluence Cody plugin',
        modelDescription: 'Search over your confluence wiki. Add your company knowledge to Cody.',
        search: (query: string) => [
            {
                title: 'People Operations',
                url: 'https://confluence.sourcegraph.com/handbook/people-operations',
                type: ['wiki page'],
                content:
                    "People Operations is a team that focuses on the people that work at Sourcegraph. We're responsible for hiring, onboarding, and supporting our employees. We also work on initiatives to make Sourcegraph a great place to work, such as our diversity and inclusion efforts, and our employee resource groups.",
            },
        ],
    },
    {
        name: 'Jira Cody plugin',
        modelDescription: 'Search over your Jira tickets. Add your company knowledge to Cody.',
        search: (query: string) => [
            {
                title: 'Incorrect post signup redirect',
                url: 'https://jira.sourcegraph.com/browse/DEV-123',
                type: ['jira_ticket'],
                content:
                    "When redirecting after user signs up, we incorrectly use 'redirect_to' instead of 'return_to' parameter in the URL. This causes the user to be redirected to the wrong page after signup.",
            },
        ],
    },
    {
        name: 'Github Cody plugin',
        modelDescription: 'Search over your Github repositories. Add your company knowledge to Cody.',
        search: (query: string) => [
            {
                title: 'feat(plugins): add cody client-side plugins',
                type: ['pull_request'],
                url: 'https://github.com/sourcegraph/sourcegraph/pull/123',
                content:
                    "This PR adds the client-side plugin infrastructure for Cody. It's not yet wired up to the server.",
                comments: [
                    {
                        author: 'John Smith',
                        status: 'approved',
                        content: "Looks good to me! Let's ship it.",
                    },
                    {
                        author: 'Jane Doe',
                        status: 'requested_changes',
                        content: "I'm not sure about this. Can we add a test?",
                    },
                ],
            },
        ],
    },
]

export const getPluginContextData = (query: string): Message[] => {
    let text = ''
    for (const plugin of PLUGINS) {
        // todo: run in sandbox environment
        const items = plugin.search(query)
        if (items.length === 0) {
            continue
        }
        text += `from ${plugin.name} data source
            \`\`\`json
            ${JSON.stringify(items)}
            \`\`\``
    }

    if (!text) {
        // skip if no results
        return []
    }

    return [
        {
            speaker: 'human',
            text: 'I have following results from different data sources:\n' + text,
        },
        {
            speaker: 'assistant',
            text: 'Understood, I have context about the above data and can use it to answer your questions.',
        },
    ]
}
