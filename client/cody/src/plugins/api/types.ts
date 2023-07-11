export interface IPluginFunction {
    name: string
    description: string
    parameters: {
        type: 'object'
        properties: {
            [key: string]: {
                type: 'string' | 'number' | 'boolean'
                enum?: string[]
                description?: string
            }
        }
        description?: string
        required?: string[]
    }
    handler: (parameters: IPluginFunctionParameters) => Promise<IPluginFunctionOutput[]>
}

export interface IPluginFunctionOutput {
    url: string
    [key: string]: any
}

export type IPluginFunctionDescriptor = Omit<IPluginFunction, 'handler'>

export type IPluginFunctionParameters = Record<string, string | number | boolean>

export type IPluginFunctionCallDescriptor = [IPluginFunction, IPluginFunctionParameters]

export interface IPluginFunctionChosenDescriptor {
    name: string
    parameters: IPluginFunctionParameters
}

export interface IPlugin {
    name: string
    description: string
    dataSources: IPluginFunction[]
}
