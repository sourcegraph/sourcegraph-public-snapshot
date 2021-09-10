import type MonacoWebpackPlugin from 'monaco-editor-webpack-plugin'

/**
 * Configuration for https://github.com/microsoft/monaco-editor-webpack-plugin.
 */
export const MONACO_LANGUAGES_AND_FEATURES: Required<
    Pick<NonNullable<ConstructorParameters<typeof MonacoWebpackPlugin>[0]>, 'languages' | 'features'>
> = {
    languages: ['json', 'yaml'],
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
