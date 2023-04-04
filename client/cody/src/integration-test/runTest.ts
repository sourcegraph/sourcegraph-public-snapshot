import * as path from 'path'

import { runTests } from '@vscode/test-electron'

import * as mockServer from './mock-server'

async function main(): Promise<void> {
    // Set this environment variable so the extension exposes hooks to the test runner.
    process.env.CODY_TESTING = 'true'

    // When run, this script's filename is `client/cody/out/src/integration-test/runTest.js`, so
    // __dirname is derived from that path, not this file's source path.
    const codyRoot = path.resolve(__dirname, '..', '..', '..')

    // The test workspace is not copied to out/ during the TypeScript build, so we need to refer to
    // it in the src/ dir.
    const testWorkspacePath = path.resolve(codyRoot, 'src', 'integration-test', 'workspace')

    try {
        // The directory containing the extension's package.json, passed to
        // --extensionDevelopmentPath.
        const extensionDevelopmentPath = codyRoot

        // The path to the test runner script, passed to --extensionTestsPath.
        const extensionTestsPath = path.resolve(codyRoot, 'out', 'src', 'integration-test', 'index')

        const launchArgs = [
            testWorkspacePath,
            '--disable-extensions', // disable other extensions
        ]

        // Download VS Code, unzip it, and run the integration test.
        await mockServer.run(() => runTests({ extensionDevelopmentPath, extensionTestsPath, launchArgs }))
    } catch {
        console.error('Failed to run tests')
        process.exit(1)
    }
}

// eslint-disable-next-line @typescript-eslint/no-floating-promises
main()
