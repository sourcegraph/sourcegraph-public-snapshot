import fs from 'fs'
import path from 'path'

import { parse } from '@sqs/jsonc-parser'

import { ROOT_PATH, MONACO_EDITOR_PATH } from '@sourcegraph/build-config'

export const monacoEditorPath = MONACO_EDITOR_PATH
export const storybookWorkspacePath = path.resolve(ROOT_PATH, 'client/storybook')
export const dllBuildPath = path.resolve(storybookWorkspacePath, 'assets/dll-bundle')
export const dllBundleManifestPath = path.resolve(dllBuildPath, 'dll-bundle.manifest.json')

// eslint-disable-next-line no-sync
export const readJsonFile = (path: string): unknown => parse(fs.readFileSync(path, 'utf-8')) as unknown

export const dllPluginConfig = {
    context: dllBuildPath,
    name: 'dll_lib',
    path: path.resolve(dllBuildPath, 'dll-plugin.manifest.json'),
}
