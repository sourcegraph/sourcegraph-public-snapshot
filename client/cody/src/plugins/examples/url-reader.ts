import { IPlugin, IPluginFunctionOutput, IPluginFunctionParameters } from '../api/types'

async function fetchURL(url: string): Promise<IPluginFunctionOutput> {
    // Use the fetch API to get the webpage content
    const response = await fetch(url)

    // Check if the request was successful
    if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`)
    }

    // Parse the webpage content as text
    const text = await response.text()

    // Return the web page content
    const result = {
        url,
        body: text.slice(0, 1000),
    }

    return result
}

export const urlReaderPlugin: IPlugin = {
    name: 'URL Reader',
    description: 'Get the content of a web page by URL.',
    dataSources: [
        {
            name: 'get_web_page_content',
            description: 'Get the content of a web page by URL.',
            parameters: {
                type: 'object',
                properties: {
                    url: {
                        type: 'string',
                        description: 'The URL of the web page to get the content of.',
                    },
                },
                required: ['url'],
            },
            handler: async (parameters: IPluginFunctionParameters): Promise<IPluginFunctionOutput[]> => {
                if (typeof parameters?.url === 'string') {
                    return fetchURL(parameters.url).then(page => [page])
                }
                return Promise.reject(new Error('Invalid parameters'))
            },
        },
    ],
}
