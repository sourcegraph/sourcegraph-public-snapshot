import path from 'path'

import * as esbuild from 'esbuild'
import { EditorFeature, featuresArr } from 'monaco-editor-webpack-plugin/out/features'
import { EditorLanguage, languagesArr } from 'monaco-editor-webpack-plugin/out/languages'

import { MONACO_LANGUAGES_AND_FEATURES } from '@sourcegraph/build-config'

import { ROOT_PATH } from '../utils'

const monacoModulePath = (modulePath: string): string =>
    require.resolve(path.join('monaco-editor/esm', modulePath), {
        paths: [path.join(ROOT_PATH, 'node_modules')],
    })

/**
 * An esbuild plugin that omits some unneeded features and languages from monaco-editor when
 * bundling, to reduce bundle size and speed up builds. Similar to
 * https://github.com/microsoft/monaco-editor-webpack-plugin.
 */
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

        // For omitted features and languages, treat their modules as empty files.
        //
        // TODO(sqs): This is different from how
        // https://github.com/microsoft/monaco-editor-webpack-plugin does it. The
        // monaco-editor-webpack-plugin approach relies on injecting a different central module
        // file, rather than zeroing out each feature/language module. Our approach necessitates the
        // ALWAYS_ENABLED_FEATURES hack above. Our approach is fine for when esbuild is still an
        // optional prototype build method for local dev, but this implementation should be fixed if
        // we switch to esbuild by default.
        build.onLoad({ filter }, () => ({ contents: '', loader: 'js' }))
    },
})
