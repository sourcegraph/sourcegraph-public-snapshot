import { IPlugin, IPluginFunctionOutput, IPluginFunctionParameters } from '../api/types'

const documents = [
    {
        title: 'People Operations',
        url: 'https://confluence.sourcegraph.com/handbook/people-operations',
        type: ['wiki page'],
        content:
            "People Operations is a team that focuses on the people that work at Sourcegraph. We're responsible for hiring, onboarding, and supporting our employees. We also work on initiatives to make Sourcegraph a great place to work, such as our diversity and inclusion efforts, and our employee resource groups.",
    },
]

export const confluencePlugin: IPlugin = {
    name: 'Confluence Cody plugin',
    description:
        'Search Confluence pages where company shared knowledge is stored as a wiki. Use this to find out how to do something, what something means, or to get a better understanding of a topic.',
    dataSources: [
        {
            name: 'search_confluence_wiki_pages',
            description:
                'Search Confluence pages where company shared knowledge is stored as a wiki. Use this to find out how to do something, what something means, or to get a better understanding of a topic.',
            parameters: {
                type: 'object',
                properties: {
                    titleQuery: {
                        type: 'string',
                        description: 'Query by page title',
                    },
                    pageContentQuery: {
                        type: 'string',
                        description: 'Query by page content',
                    },
                },
                required: ['pageContentQuery', 'titleQuery'],
            },
            handler: (parameters: IPluginFunctionParameters): Promise<IPluginFunctionOutput[]> =>
                Promise.resolve(documents),
        },
    ],
}
