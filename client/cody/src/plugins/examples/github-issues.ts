import { pick } from 'lodash'

import { IPlugin, IPluginFunctionOutput, IPluginFunctionParameters } from '../api/types'

// TODO: env variables or pluging configuration
// TODO: support OAuth tokens as well
const apiToken = 'API_TOKEN'
const org = 'sourcegraph'

const baseURL = 'https://api.github.com/'
const path = '/search/issues'

interface User {
    name: string | null
    login: string
    html_url: string
}

interface Item {
    title: string
    body: string
    html_url: string
    assignee: User | null
    user: User | null
    created_at: string
    updated_at: string
    state: string
}

const searchGitHub = async (query: string): Promise<any> => {
    const url = new URL(path, baseURL)
    // TODO: what is a good limit of results here?
    url.searchParams.set('per_page', '2')
    url.searchParams.set('q', `${query} org:${org}`)
    const opts = {
        method: 'GET',
        headers: {
            Authorization: apiToken,
            'Content-Type': 'application/json',
        },
    }

    try {
        const rawResponse = await fetch(url, opts)
        const response = await rawResponse.json()

        const items = response?.items as Item[]
        return items.map(item => {
            const title = item.title.slice(0, 180) // up to 180 characters
            const body = removeHtmlTags(item.body).slice(0, 300) // up to 300 characters
            const user =
                item.user != null
                    ? {
                          name: item.user.name,
                          login: item.user.login,
                          url: item.user.html_url,
                      }
                    : null
            const assignee =
                item.assignee != null
                    ? {
                          name: item.assignee.name,
                          login: item.assignee.login,
                          url: item.assignee.html_url,
                      }
                    : null
            return {
                ...pick(item, ['created_at', 'updated_at', 'state']),
                title,
                body,
                user,
                assignee,
                url: item.html_url,
            }
        })
    } catch (error) {
        // Handle and log any errors
        console.error('Error in searchGitHub:', error)
        throw error
    }
}

// todo: add isEnabled check function
export const githubIssuesPlugin: IPlugin = {
    name: 'GitHub Issues Cody plugin',
    description:
        'Search GitHub Issues and pull requests. Use this to understand current code problems, feature implementations, their rationale, and to gain deeper insights into various topics.',
    dataSources: [
        {
            name: 'search_github_issues',
            description:
                'Search GitHub Issues and pull requests. Use this to understand current code problems, feature implementations, their rationale, and to gain deeper insights into various topics.',
            parameters: {
                type: 'object',
                properties: {
                    query: {
                        type: 'string',
                        description: 'Query, uses github search query format',
                    },
                },
                required: ['query'],
            },
            handler: async (parameters: IPluginFunctionParameters): Promise<IPluginFunctionOutput[]> => {
                const { query } = parameters

                if (typeof query !== 'string') {
                    return Promise.reject(new Error('Invalid parameters'))
                }

                return searchGitHub(query)
            },
        },
    ],
}

// This function is meant to sanitize confluence content that's semi-html:
// <h1> hello </h1> and @@@h1@@@ hello @@@endh1@@@ will both result in `hello`.
function removeHtmlTags(input: string): string {
    return input.replace(/<[^>]+>|@@@[^@]+@@@/g, '')
}
