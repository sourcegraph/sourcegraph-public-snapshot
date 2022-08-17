declare module 'esbuild-plugin-elm' {
    function ElmPlugin(config: ElmPluginConfig): esbuild.Plugin

    export default ElmPlugin
}

interface ElmPluginConfig {
    debug?: boolean
    optimize?: boolean
    pathToElm?: string
    clearOnWatch?: boolean
    cwd?: string
}
