import { IPlugin, IPluginFunctionOutput, IPluginFunctionParameters } from '../api/types'

export const weatherPlugin: IPlugin = {
    name: 'Weather Cody plugin',
    description: "Search weather. Use this to find out what's the weather today, tomorrow, or next week.",
    dataSources: [
        {
            name: 'get_current_weather',
            description: 'Get the current weather in a given location',
            parameters: {
                type: 'object',
                properties: {
                    location: {
                        type: 'string',
                        description: 'The city and state, e.g. San Francisco, CA',
                    },
                    unit: { type: 'string', enum: ['celsius', 'fahrenheit'] },
                },
                required: ['location'],
            },
            handler: (parameters: IPluginFunctionParameters): Promise<IPluginFunctionOutput[]> => Promise.resolve([]),
        },
    ],
}
