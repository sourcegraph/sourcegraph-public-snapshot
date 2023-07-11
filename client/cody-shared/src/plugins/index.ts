// TODO: move this implementation to cody-shared
export interface PluginContextData {
    url: string
}
export interface Plugin {
    name: string
    modelDescription: string
    // todo: support different argument schemas
    search: (query: string) => PluginContextData[]
}
