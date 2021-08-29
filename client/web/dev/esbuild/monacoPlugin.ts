import fs from 'fs'
import os from 'os'
import path from 'path'

import * as esbuild from 'esbuild'
import { EditorFeature, featuresArr } from 'monaco-editor-webpack-plugin/out/features'
import { EditorLanguage, languagesArr } from 'monaco-editor-webpack-plugin/out/languages'

import { MONACO_LANGUAGES_AND_FEATURES } from '../webpack/monacoWebpack'

// TODO(sqs): this was found to NOT be faster than not using it

const monacoModulePath = (modulePath: string): string =>
    require.resolve(path.join('monaco-editor/esm', modulePath), {
        paths: [path.join(__dirname, '../../../../node_modules')],
    })

export const monacoPlugin = ({
    languages,
    features,
}: Required<typeof MONACO_LANGUAGES_AND_FEATURES>): esbuild.Plugin => ({
    name: 'monaco',
    setup: build => {
        for (const feature of features) {
            if (feature.startsWith('!')) {
                throw new Error('negated features (starting with "!") are not supported')
            }
        }

        // Some feature exclusions don't work because their module exports a symbol needed by
        // another feature.
        const ALWAYS_ENABLED_FEATURES = new Set<EditorFeature>(['snippets'])

        const skipLanguageModules = languagesArr
            .filter(({ label }) => !languages.includes(label as EditorLanguage))
            .flatMap(({ entry }) => entry || [])
        const skipFeatureModules = featuresArr
            .filter(
                ({ label }) =>
                    !features.includes(label as EditorFeature) && !ALWAYS_ENABLED_FEATURES.has(label as EditorFeature)
            )
            .flatMap(({ entry }) => entry || [])

        const skipModulePaths = [...skipLanguageModules, ...skipFeatureModules].map(monacoModulePath)
        const filter = new RegExp(`^(${skipModulePaths.join('|')})$`)

        build.onLoad({ filter }, () => ({ contents: '', loader: 'js' }))
    },
})
