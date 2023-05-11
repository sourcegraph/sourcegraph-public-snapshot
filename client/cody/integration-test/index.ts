import * as path from 'path'

import glob from 'glob'
import Mocha from 'mocha'

export function run(): Promise<void> {
    const mocha = new Mocha({
        ui: 'tdd',
        color: true,
        timeout: 15000,
        grep: process.env.TEST_PATTERN ? new RegExp(process.env.TEST_PATTERN, 'i') : undefined,
    })

    const testsRoot = __dirname

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
