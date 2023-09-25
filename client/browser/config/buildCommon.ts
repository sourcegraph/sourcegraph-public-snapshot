import path from 'path'

import { ROOT_PATH } from '@sourcegraph/build-config'

const IS_BAZEL = !!(process.env.JS_BINARY__TARGET || process.env.BAZEL_BINDIR)

export const browserWorkspacePath = path.resolve(IS_BAZEL ? process.cwd() : ROOT_PATH, 'client/browser')
const browserSourcePath = path.resolve(browserWorkspacePath, 'src')

function resolveSourcePath(sourcePath: string): string {
    if (IS_BAZEL) {
        sourcePath = sourcePath.replace(/\.tsx?$/, '.js')
    }
    return path.resolve(browserSourcePath, sourcePath)
}

/**
 * Build entrypoints for the browser extension.
 */
export const entrypoints: Record<string, string> = {
    // Browser extension
    background: resolveSourcePath('browser-extension/scripts/backgroundPage.main.ts'),
    inject: resolveSourcePath('browser-extension/scripts/contentPage.main.ts'),
    options: resolveSourcePath('browser-extension/scripts/optionsPage.main.tsx'),
    'after-install': resolveSourcePath('browser-extension/scripts/afterInstallPage.main.tsx'),

    // Common native integration entry point (Gitlab, Bitbucket)
    integration: resolveSourcePath('native-integration/integration.main.ts'),
    // Phabricator-only native integration entry point
    phabricator: resolveSourcePath('native-integration/phabricator/integration.main.ts'),

    // Styles
    style: resolveSourcePath('app.scss'),
    'branded-style': resolveSourcePath('branded.scss'),
}
