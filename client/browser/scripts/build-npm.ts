import * as path from 'path'

import latestVersion from 'latest-version'
import { writeFile } from 'mz/fs'
import * as semver from 'semver'
import signale from 'signale'

import { copyInlineExtensions } from './tasks'

export const packagePath = path.resolve(__dirname, '..', 'build', 'integration')

/**
 * Build a new native integration to npm package
 */
export async function buildNpm(bumpVersion?: boolean): Promise<void> {
    const name = '@sourcegraph/code-host-integration'
    // Bump version
    let version = '0.0.0'
    if (bumpVersion) {
        try {
            const currentVersion = await latestVersion(name)
            signale.info(`Current version is ${currentVersion}`)
            version = semver.inc(currentVersion, 'patch')!
        } catch (error) {
            if (error && error.name === 'PackageNotFoundError') {
                signale.info('Package is not released yet')
            } else {
                throw error
            }
        }
    }

    const packageJson = {
        name,
        version,
        license: 'Apache-2.0',
        repository: {
            type: 'git',
            url: 'https://github.com/sourcegraph/sourcegraph',
            directory: 'browser',
        },
    }
    signale.info(`New version is ${packageJson.version}`)
    // Write package.json
    await writeFile(path.join(packagePath, 'package.json'), JSON.stringify(packageJson, null, 2))

    copyInlineExtensions(packagePath)
}
