import execa from 'execa'
import signale from 'signale'

import { buildNpm, packagePath } from './build-npm'

/**
 * Publish the native integration to npm
 */
async function main(): Promise<void> {
    await buildNpm(true)
    if (!process.env.CI) {
        signale.warn('Not running in CI, aborting')
        return
    }
    // Publish
    await execa('npm', ['publish', '--access', 'public'], { cwd: packagePath, stdio: 'inherit' })
}
main().catch(error => {
    process.exitCode = 1
    console.error(error)
})
