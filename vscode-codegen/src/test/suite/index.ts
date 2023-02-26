import * as path from 'path'

import glob from 'glob'
import Mocha from 'mocha'

export function run(): Promise<void> {
	// Create the mocha test
	const mocha = new Mocha({
		ui: 'tdd',
		color: true,
	})

	const testsRoot = path.resolve(__dirname, '..')

	return new Promise((c, e) => {
		glob('**/**.test.js', { cwd: testsRoot }, (err, files) => {
			if (err) {
				return e(err)
			}

			// Add files to the test suite
			files.forEach(f => mocha.addFile(path.resolve(testsRoot, f)))

			try {
				// Run the mocha test
				mocha.run(failures => {
					if (failures > 0) {
						e(new Error(`${failures} tests failed.`))
					} else {
						c()
					}
				})
			} catch (error) {
				console.error(error)
				e(error)
			}
		})
	})
}
