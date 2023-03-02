import * as http from 'http'
import * as path from 'path'
import * as ws from 'ws'

import { runTests } from '@vscode/test-electron'

// Runs a stub Cody service for testing.
async function runMockServer(callback: () => Promise<any>) {
	const serverPort = 49300
	const embeddingPort = 49301
	// TODO: Extend these servers to support expectations.
	const socketServer = new ws.WebSocketServer({
		port: serverPort,
	})
	await new Promise(resolve => {
		socketServer.on('connection', socket => {
			socket.on('data', data => {
				socket.send('{"kind": "response:complete", "message": "hello, world"}')
			})
		})
		socketServer.on('listening', resolve)
	})

	const embeddingServer = http.createServer()
	embeddingServer.on('request', (request, response) => {
		response.statusCode = 200
		response.setHeader('content-type', 'text/plain')
		response.write('hello, world')
		response.end()
	})
	embeddingServer.listen(embeddingPort)

	await callback()

	socketServer.close()
	embeddingServer.close()
}

async function main() {
	// Set this environment variable so the extension exposes hooks to
	// the test runner.
	process.env['CODY_TESTING'] = 'true'

	try {
		// The folder containing the Extension Manifest package.json
		// Passed to `--extensionDevelopmentPath`
		const extensionDevelopmentPath = path.resolve(__dirname, '../../')

		// The path to test runner
		// Passed to --extensionTestsPath
		const extensionTestsPath = path.resolve(__dirname, './suite/index')

		const launchArgs = [
			path.resolve(__dirname, '../../src/test/workspace'), // Test workspace
			'--disable-extensions', // Disable other extensions
		]

		// Download VS Code, unzip it and run the integration test
		await runMockServer(() => runTests({ extensionDevelopmentPath, extensionTestsPath, launchArgs }))
	} catch {
		console.error('Failed to run tests')
		process.exit(1)
	}
}

main()
