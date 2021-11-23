/* eslint-disable unicorn/prevent-abbreviations */
/* eslint-disable @typescript-eslint/consistent-type-assertions */
import { Config, ConfigOptions } from 'karma'

import webpackConfig from './webpack.test.config'

// eslint-disable-next-line import/no-default-export
export default (config: Config): void => {
    config.set({
        frameworks: ['mocha', 'chai', 'sinon'],

        files: [
            {
                pattern: 'src/**/*.ts',
                watched: false,
                included: true,
                served: true,
            },
        ],
        preprocessors: {
            'src/**/*.ts?(x)': ['webpack', 'sourcemap'],
        },

        // Ignore the npm package entry point
        exclude: ['src/index.ts'],

        // karma-webpack doesn't change the file extensions so we just need to tell karma what these extensions mean.
        mime: {
            'text/x-typescript': ['ts'],
        },

        webpack: webpackConfig,

        webpackMiddleware: {
            stats: 'errors-only',
            bail: true,
        },

        browsers: ['Chrome', 'Firefox'],
        customLaunchers: {
            ChromeHeadlessNoSandbox: {
                base: 'ChromeHeadless',
                // CI runs as root so we need to disable sandbox
                flags: ['--no-sandbox', '--disable-setuid-sandbox'],
            },
        },

        reporters: ['mocha', 'coverage-istanbul'],
        mochaReporter: {
            showDiff: true,
        },
        client: {
            mocha: {
                timeout: 6000,
            },
        },
        coverageIstanbulReporter: {
            reports: ['json'],
            fixWebpackSourcePaths: true,
        },
    } as ConfigOptions)
}
