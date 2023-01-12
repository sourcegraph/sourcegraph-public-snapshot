import { realpathSync } from 'fs'
import path from 'path'

// NOTE: use fs.realpathSync() in addition to path.resolve() to resolve
// symlinks to the real path. This is required for webpack plugins using
// `include: [...file path...]` when the file path contains symlinks such
// as when using pnpm.
export function resolveWithSymlink(...args: string[]): string {
    const resolvedPath = path.resolve(...args)

    try {
        return realpathSync(resolvedPath)
    } catch {
        return resolvedPath
    }
}

export const ROOT_PATH = resolveWithSymlink(__dirname, '../../../')
export const WORKSPACES_PATH = resolveWithSymlink(ROOT_PATH, 'client')
export const NODE_MODULES_PATH = resolveWithSymlink(ROOT_PATH, 'node_modules')
export const MONACO_EDITOR_PATH = resolveWithSymlink(NODE_MODULES_PATH, 'monaco-editor')
export const STATIC_ASSETS_PATH = resolveWithSymlink(ROOT_PATH, 'ui/assets')
export const STATIC_INDEX_PATH = resolveWithSymlink(STATIC_ASSETS_PATH, 'index.html')
