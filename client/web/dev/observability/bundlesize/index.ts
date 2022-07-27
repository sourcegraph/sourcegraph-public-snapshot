/* eslint-disable @typescript-eslint/no-var-requires, @typescript-eslint/no-require-imports */
/**
 * Build web application using:
 * ENTERPRISE=1 NODE_ENV=production DISABLE_TYPECHECKING=true WEBPACK_USE_NAMED_CHUNKS=true yarn build-web
 *
 * Upload bundlesize information to Honeycomb:
 * HONEYCOMB_API_KEY=XXX yarn workspace @sourcegraph/web run bundlesize:upload
 *
 * Check out collected data in Honeycomb! 🧠
 */
import { execSync } from 'child_process'
import { statSync } from 'fs'
import path from 'path'

import { SemanticResourceAttributes } from '@opentelemetry/semantic-conventions'
import { cleanEnv, bool, str } from 'envalid'
import glob from 'glob'

import { STATIC_ASSETS_PATH, WORKSPACES_PATH } from '@sourcegraph/build-config'

import { BUILDKITE_INFO, SDK_INFO } from '../constants'
import { libhoneySDK } from '../sdk'

const environment = cleanEnv(process.env, {
    ENTERPRISE: bool({ default: false }),
    NODE_ENV: str({ choices: ['development', 'production'] }),
})

interface BundleSizeConfig {
    files: {
        path: string
        maxSize: string
    }[]
}

// eslint-disable-next-line import/no-dynamic-require
const bundleSizeConfig = require(path.join(WORKSPACES_PATH, 'web/bundlesize.config')) as BundleSizeConfig

/**
 * Get a list of files specified by bundlesize config glob and their respective sizes.
 */
const bundleStats = bundleSizeConfig.files.reduce<Record<string, number>>((result, file) => {
    const filePaths = glob.sync(file.path)

    const fileStats = filePaths.reduce(
        (fileStats, filePath) => ({
            ...fileStats,
            [filePath.replace(STATIC_ASSETS_PATH, '')]: statSync(filePath).size,
        }),
        {}
    )

    return {
        ...result,
        ...fileStats,
    }
}, {})

const commit = execSync('git rev-parse HEAD').toString().trim()

/**
 * Log every file size as a separate event to Honeycomb.
 */
for (const [fileName, size] of Object.entries(bundleStats)) {
    libhoneySDK.sendNow({
        name: 'bundlesize',
        [SemanticResourceAttributes.SERVICE_NAME]: 'bundlesize',
        [SemanticResourceAttributes.SERVICE_NAMESPACE]: 'web',
        [SemanticResourceAttributes.SERVICE_VERSION]: commit,

        'bundle.file.name': fileName,
        'bundle.file.size': size,
        'bundle.enterprise': environment.ENTERPRISE,
        'bundle.env': environment.NODE_ENV,

        ...SDK_INFO,
        ...BUILDKITE_INFO,
    })
}
