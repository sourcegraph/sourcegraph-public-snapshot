import MonacoWebpackPlugin from 'monaco-editor-webpack-plugin'
import { WebpackPluginInstance, RuleSetRule } from 'webpack'

import { MONACO_EDITOR_PATH } from '../paths'

// CSS rule for monaco-editor and other external plain CSS (skip SASS and PostCSS for build perf)
export const getMonacoCSSRule = (): RuleSetRule => ({
    test: /\.css$/,
    include: [MONACO_EDITOR_PATH],
    use: ['style-loader', { loader: 'css-loader' }],
})

// TTF rule for monaco-editor
export const getMonacoTTFRule = (): RuleSetRule => ({
    test: /\.ttf$/,
    include: [MONACO_EDITOR_PATH],
    type: 'asset/resource',
})

/**
 * Configuration for https://github.com/microsoft/monaco-editor-webpack-plugin.
 */
export const MONACO_LANGUAGES_AND_FEATURES: Required<
    Pick<
        NonNullable<ConstructorParameters<typeof MonacoWebpackPlugin>[0]>,
        'languages' | 'features' | 'customLanguages'
    >
> = {
    languages: ['json', 'yaml'],
    customLanguages: [
        {
            label: 'yaml',
            entry: 'monaco-yaml/lib/esm/monaco.contribution',
            worker: { id: 'vs/language/yaml/yamlWorker', entry: 'monaco-yaml/lib/esm/yaml.worker' },
        },
    ],
    features: [
        'bracketMatching',
        'clipboard',
        'coreCommands',
        'cursorUndo',
        'find',
        'format',
        'hover',
        'inPlaceReplace',
        'iPadShowKeyboard',
        'links',
        'suggest',
    ],
}

export const getMonacoWebpackPlugin = (): WebpackPluginInstance =>
    new MonacoWebpackPlugin(MONACO_LANGUAGES_AND_FEATURES)
