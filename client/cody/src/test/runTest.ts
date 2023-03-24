import * as path from 'path'

import { runTests } from '@vscode/test-electron'

import * as mockServer from './mock-server'

async function main(): Promise<void> {
    // Set this environment variable so the extension exposes hooks to
    // the test runner.
    process.env.CODY_TESTING = 'true'

    try {
        // The folder containing the Extension Manifest package.json
        // Passed to `--extensionDevelopmentPath`
        const extensionDevelopmentPath = path.resolve(__dirname, '../../../../')

        // The path to test runner
        // Passed to --extensionTestsPath
        const extensionTestsPath = path.resolve(__dirname, './suite/index')

        const launchArgs = [
            path.resolve(__dirname, '../../../../src/test/workspace'), // Test workspace
            '--disable-extensions', // Disable other extensions
        ]

        // Download VS Code, unzip it and run the integration test
        await mockServer.run(() => runTests({ extensionDevelopmentPath, extensionTestsPath, launchArgs }))
    } catch {
        console.error('Failed to run tests')
        process.exit(1)
    }
}

main()
