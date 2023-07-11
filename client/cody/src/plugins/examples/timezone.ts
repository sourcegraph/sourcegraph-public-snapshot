import { IPlugin, IPluginFunctionOutput, IPluginFunctionParameters } from '../api/types'

import { fetchAPINinjas } from './lib/fetch-api-ninjas'

export const timezonePlugin: IPlugin = {
    name: 'Timezone Cody plugin',
    description: 'Search timezone. Use this to find out what is the timezone in different cities.',
    dataSources: [
        {
            name: 'get_city_time',
            description: 'Get the current time for a given city',
            parameters: {
                type: 'object',
                properties: {
                    city: {
                        type: 'string',
                        description: 'A valid full city name to get the weather for, e.g San Francisco',
                    },
                },
                required: ['city'],
            },
            handler: (parameters: IPluginFunctionParameters): Promise<IPluginFunctionOutput[]> => {
                if (typeof parameters?.city !== 'string') {
                    return Promise.reject(new Error('Invalid parameters'))
                }
                const url = 'https://api.api-ninjas.com/v1/worldtime?city=' + parameters.city
                return fetchAPINinjas(url).then(async response => {
                    if (!response.ok) {
                        return [
                            {
                                url,
                                error: 'Could not fetch timezone data',
                            },
                        ]
                    }
                    const json = await response.json()
                    return [
                        {
                            url: '',
                            city: parameters.city,
                            current_datetime: json.datetime,
                            current_day_of_week: json.day_of_week,
                            timezone: json.timezone,
                        },
                    ]
                })
            },
        },
    ],
}
