import { spawnSync } from 'child_process'
import fs from 'fs'
import path from 'path'

import signale from 'signale'

import { ROOT_PATH } from '@sourcegraph/build-config'

import { dllPluginConfig, storybookWorkspacePath } from '../webpack.config.common'

// Build DLL bundle with `pnpm build:dll-bundle` if it's not available.
export const ensureDllBundleIsReady = (): void => {
    signale.start(`Checking if DLL bundle is available: ${path.relative(ROOT_PATH, dllPluginConfig.path)}`)

    // eslint-disable-next-line no-sync
    if (!fs.existsSync(dllPluginConfig.path)) {
        signale.warn('DLL bundle not found!')
        signale.await('Building new DLL bundle with `pnpm build:dll-bundle`')

        spawnSync('pnpm', ['build:dll-bundle'], {
            stdio: 'inherit',
            cwd: storybookWorkspacePath,
        })
    }

    signale.success('DLL bundle is ready!')
}
