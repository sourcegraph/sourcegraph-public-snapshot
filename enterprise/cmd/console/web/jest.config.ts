/// <reference types="node" />

import type { Config } from '@jest/types'

/** @type {jest.InitialOptions} */
const config: Config.InitialOptions = {
    collectCoverage: !!process.env.CI,
    coverageDirectory: '<rootDir>/coverage',
    coveragePathIgnorePatterns: [/\.test\.tsx?$/.source],
    roots: ['<rootDir>/src'],
    setupFilesAfterEnv: ['./src/setupTests.ts'],

    // By default, don't clutter `yarn test --watch` output with the full coverage table. To see it, use the
    // `--coverageReporters text` jest option.
    coverageReporters: ['json', 'lcov', 'text-summary'],
}

// eslint-disable-next-line import/no-default-export
export default config
