import fs from 'fs'
import path from 'path'

import { parse } from '@sqs/jsonc-parser'
import { RuleSetUseItem } from 'webpack'

export const rootPath = path.resolve(__dirname, '../../..')
export const nodeModulesPath = path.resolve(rootPath, 'node_modules')
export const monacoEditorPath = path.resolve(nodeModulesPath, 'monaco-editor')
export const storybookWorkspacePath = path.resolve(rootPath, 'client/storybook')
export const dllBuildPath = path.resolve(storybookWorkspacePath, 'assets/dll-bundle')
export const dllBundleManifestPath = path.resolve(dllBuildPath, 'dll-bundle.manifest.json')

// eslint-disable-next-line no-sync
export const readJsonFile = (path: string): unknown => parse(fs.readFileSync(path, 'utf-8')) as unknown

export const getBasicCSSLoader = (): RuleSetUseItem => ({
    loader: 'css-loader',
    options: { url: false },
})

export const dllPluginConfig = {
    context: dllBuildPath,
    name: 'dll_lib',
    path: path.resolve(dllBuildPath, 'dll-plugin.manifest.json'),
}
