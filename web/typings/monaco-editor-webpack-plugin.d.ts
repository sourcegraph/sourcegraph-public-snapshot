declare module 'monaco-editor-webpack-plugin' {
    import { Plugin } from 'webpack'
    interface MonacoEditorWebpackPluginOptions {
        output?: string
        languages?: string[]
        features?: string[]
    }
    class MonacoEditorWebpackPlugin extends Plugin {
        constructor(options: MonacoEditorWebpackPluginOptions)
    }
    export = MonacoEditorWebpackPlugin
}
