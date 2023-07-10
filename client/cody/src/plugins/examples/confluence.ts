import { IPlugin, IPluginFunctionOutput, IPluginFunctionParameters } from '../api/types'

// todo: use env variables
const email = 'EMAIL'
const apiToken =
    'API_TOKEN'

const base_url = 'https://sourcegraph-source.atlassian.net'

const searchWiki = (query: string): Promise<any> =>
    fetch(`${base_url}/wiki/rest/api/search?cql=${encodeURIComponent(`text ~ "${query}"`)}`, {
        method: 'GET',
        headers: {
            Authorization: 'Basic ' + btoa(email + ':' + apiToken),
            'Content-Type': 'application/json',
        },
    })
        .then(response => response.json())
        .then(
            json =>
                json?.results as {
                    excerpt: string
                    title: string
                    url: string
                }[]
        )
        .then(items => items.map(({ excerpt, title, url }) => ({ excerpt, title, url: `${base_url}/${url}` })))

// todo: add isEnabled check function
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
                    query: {
                        type: 'string',
                        description: 'Query by page title or content',
                    },
                },
                required: ['query'],
            },
            handler: async (parameters: IPluginFunctionParameters): Promise<IPluginFunctionOutput[]> => {
                const { query } = parameters

                if (typeof query === 'string') {
                    const items = await searchWiki(query)

                    return items
                }
                return Promise.reject(new Error('Invalid parameters'))
            },
        },
    ],
}
