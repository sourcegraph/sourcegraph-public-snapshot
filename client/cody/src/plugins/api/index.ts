import { ChatClient } from '@sourcegraph/cody-shared/src/chat/chat'
import { Message } from '@sourcegraph/cody-shared/src/sourcegraph-api'

import { defaultPlugins } from '../examples'

import { makePrompt } from './prompt'
import { IPluginFunctionCallDescriptor, IPluginFunctionChosenDescriptor } from './types'

export const chooseDataSources = (
    query: string,
    client: ChatClient,
    plugins = defaultPlugins
): Promise<IPluginFunctionCallDescriptor[]> => {
    const allDataSources = plugins.flatMap(plugin => plugin.dataSources)

    const messages = makePrompt(
        query,
        allDataSources.map(({ handler, ...rest }) => rest)
    )
    return new Promise<IPluginFunctionCallDescriptor[]>((resolve, reject) => {
        let lastResponse = ''
        client.chat(messages, {
            onChange: text => {
                lastResponse = text
            },
            onComplete: () => {
                try {
                    const descriptors = JSON.parse(lastResponse.trim()) as IPluginFunctionChosenDescriptor[]
                    resolve(
                        descriptors.map(
                            item =>
                                [
                                    allDataSources.find(dataSource => dataSource.name === item.name),
                                    item.parameters,
                                ] as IPluginFunctionCallDescriptor
                        )
                    )
                } catch (error) {
                    reject(new Error(`Error parsing llm intent detection response: ${error}`))
                }
            },
            onError: (error, statusCode) => {
                reject(new Error(`error: ${error}\nstatus code: ${statusCode}`))
            },
        })
    })
}

export const getContext = async (dataSourcesCallDescriptors: IPluginFunctionCallDescriptor[]): Promise<Message[]> => {
    let outputs = await Promise.all(
        dataSourcesCallDescriptors.map(async ([dataSource, parameters]) => {
            const response = await dataSource.handler(parameters)

            if (!response.length) {
                return
            }

            return `from ${dataSource.name} data source: \`\`\`json${JSON.stringify(response)}\`\`\``
        })
    )

    outputs = outputs.filter(Boolean)
    if (outputs.length === 0) {
        return []
    }

    return [
        {
            speaker: 'human',
            text: 'I have following results from different data sources:\n' + outputs.join('\n'),
        },
        {
            speaker: 'assistant',
            text: 'Understood, I have context about the above data and can use it to answer your questions.',
        },
    ]
}
