import { writeFile } from 'mz/fs'
import latestVersion from 'latest-version'
import signale from 'signale'
import * as semver from 'semver'
import execa from 'execa'

// Publish the native integration to npm

async function main(): Promise<void> {
    const name = '@sourcegraph/code-host-integration'
    // Bump version
    let version: string
    try {
        const currentVersion = await latestVersion(name)
        signale.info(`Current version is ${currentVersion}`)
        version = semver.inc(currentVersion, 'patch')!
    } catch (err) {
        if (err && err.name === 'PackageNotFoundError') {
            signale.info('Package is not released yet')
            version = '0.0.0'
        } else {
            throw err
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
    await writeFile(__dirname + '/../build/integration/package.json', JSON.stringify(packageJson, null, 2))
    if (!process.env.CI) {
        signale.warn('Not running in CI, aborting')
        return
    }
    // Publish
    await execa('npm', ['publish', '--access', 'public'], { stdio: 'inherit' })
}
main().catch(err => {
    process.exitCode = 1
    console.error(err)
})
