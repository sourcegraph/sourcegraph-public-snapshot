/* eslint-disable no-sync */
import fs from 'fs'
import path from 'path'

// TODO(bazel): drop when non-bazel removed.
const IS_BAZEL = !!(process.env.JS_BINARY__TARGET || process.env.BAZEL_BINDIR || process.env.BAZEL_TEST)

// NOTE: use fs.realpathSync() in addition to path.resolve() to resolve symlinks to the real path.
// This canonicalizes the path, which avoids potential bugs in the frontend bundler step.
function resolveWithSymlink(...args: string[]): string {
    const resolvedPath = path.resolve(...args)

    try {
        return fs.realpathSync(resolvedPath)
    } catch {
        return resolvedPath
    }
}

function resolveAssetsPath(root: string): string {
    if (IS_BAZEL && process.env.WEB_BUNDLE_PATH) {
        return resolveWithSymlink(root, process.env.WEB_BUNDLE_PATH)
    }

    return resolveWithSymlink(root, 'client/web/dist')
}

export const ROOT_PATH = IS_BAZEL ? process.cwd() : resolveWithSymlink(__dirname, '../../../')
export const WORKSPACES_PATH = resolveWithSymlink(ROOT_PATH, 'client')
export const NODE_MODULES_PATH = resolveWithSymlink(ROOT_PATH, 'node_modules')
export const MONACO_EDITOR_PATH = resolveWithSymlink(NODE_MODULES_PATH, 'monaco-editor')
export const STATIC_ASSETS_PATH = resolveAssetsPath(ROOT_PATH)
function getWorkspaceNodeModulesPaths(): string[] {
    const workspaces = fs.readdirSync(WORKSPACES_PATH)
    const nodeModulesPaths = workspaces.map(workspace => resolveWithSymlink(WORKSPACES_PATH, workspace, 'node_modules'))
    return nodeModulesPaths
}

export const WORKSPACE_NODE_MODULES_PATHS = getWorkspaceNodeModulesPaths()
