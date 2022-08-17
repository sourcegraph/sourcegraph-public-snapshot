import path from 'path'

export const ROOT_PATH = path.resolve(__dirname, '../../../')
export const WORKSPACES_PATH = path.resolve(ROOT_PATH, 'client')
export const NODE_MODULES_PATH = path.resolve(ROOT_PATH, 'node_modules')
export const MONACO_EDITOR_PATH = path.resolve(NODE_MODULES_PATH, 'monaco-editor')
export const STATIC_ASSETS_PATH = path.resolve(ROOT_PATH, 'ui/assets')
export const STATIC_INDEX_PATH = path.resolve(STATIC_ASSETS_PATH, 'index.html')
