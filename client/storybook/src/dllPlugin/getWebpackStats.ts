import { spawnSync } from 'child_process'
import fs from 'fs'
import path from 'path'

import signale from 'signale'
import { StatsCompilation } from 'webpack'

import { ROOT_PATH } from '@sourcegraph/build-config'

import { readJsonFile, storybookWorkspacePath } from '../webpack.config.common'

const webpackStatsPath = path.resolve(storybookWorkspacePath, 'storybook-static/preview-stats.json')

export const ensureWebpackStatsAreReady = (): void => {
    signale.start(`Checking if Webpack stats are available: ${path.relative(ROOT_PATH, webpackStatsPath)}`)

    // eslint-disable-next-line no-sync
    if (!fs.existsSync(webpackStatsPath)) {
        signale.warn('Webpack stats not found!')
        signale.await('Building Webpack stats with `pnpm build:webpack-stats`')

        spawnSync('pnpm', ['build:webpack-stats'], {
            stdio: 'inherit',
            cwd: storybookWorkspacePath,
        })
    }

    signale.success('Webpack stats are ready!')
}

// Read Webpack stats JSON file. If it's not available use `pnpm build:webpack-stats` command to create it.
export function getWebpackStats(): StatsCompilation {
    ensureWebpackStatsAreReady()

    return readJsonFile(webpackStatsPath) as StatsCompilation
}
