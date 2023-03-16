import * as path from 'path'

import glob from 'glob'
import Mocha from 'mocha'

export function run(): Promise<void> {
    // Create the mocha test
    const mocha = new Mocha({
        ui: 'tdd',
        color: true,
    })
    // To debug tests interactively, extend this timeout.
    mocha.timeout(2000)

    const testsRoot = path.resolve(__dirname, '..')

    return new Promise((resolve, reject) => {
        glob('**/**.test.js', { cwd: testsRoot }, (err, files) => {
            if (err) {
                return reject(err)
            }

            // Add files to the test suite
            for (const file of files) {
                mocha.addFile(path.resolve(testsRoot, file))
            }

            try {
                // Run the mocha test
                mocha.run(failures => {
                    if (failures > 0) {
                        reject(new Error(`${failures} tests failed.`))
                    } else {
                        resolve()
                    }
                })
            } catch (error) {
                console.error(error)
                reject(error)
            }
        })
    })
}
