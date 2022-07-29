/**
 * The script collects web application bundlesize information from the disk and uploads it to Honeycomb.
 *
 * 1. Build web application using:
 * ENTERPRISE=1 NODE_ENV=production DISABLE_TYPECHECKING=true WEBPACK_USE_NAMED_CHUNKS=true yarn build-web
 *
 * 2. Upload bundlesize information to Honeycomb:
 * HONEYCOMB_API_KEY=XXX yarn workspace @sourcegraph/observability run bundlesize:web:upload
 *
 * 3. Check out collected data in Honeycomb! 🧠
 */
import { execSync } from 'child_process'
import path from 'path'

import { SemanticResourceAttributes } from '@opentelemetry/semantic-conventions'
import { cleanEnv, bool, str } from 'envalid'

import { WORKSPACES_PATH } from '@sourcegraph/build-config'

import { BUILDKITE_INFO, SDK_INFO } from '../constants'
import { libhoneySDK } from '../sdk'

import { getBundleSizeStats } from './getBundleSizeStats'

const environment = cleanEnv(process.env, {
    ENTERPRISE: bool({ default: false }),
    NODE_ENV: str({ choices: ['development', 'production'] }),
})

const bundleSizeStats = getBundleSizeStats(path.join(WORKSPACES_PATH, 'web/bundlesize.config'))
const commit = execSync('git rev-parse HEAD').toString().trim()
const branch = execSync('git rev-parse --abbrev-ref HEAD').toString().trim()

/**
 * Log every file size as a separate event to Honeycomb.
 */
for (const [baseFilePath, fileSizes] of Object.entries(bundleSizeStats)) {
    libhoneySDK.sendNow({
        name: 'bundlesize',
        [SemanticResourceAttributes.SERVICE_NAME]: 'bundlesize',
        [SemanticResourceAttributes.SERVICE_NAMESPACE]: 'web',
        [SemanticResourceAttributes.SERVICE_VERSION]: commit,
        'service.branch': branch,

        'bundle.file.name': baseFilePath,
        'bundle.file.size.raw': fileSizes.raw,
        'bundle.file.size.gzip': fileSizes.gzip,
        'bundle.file.size.brotli': fileSizes.brotli,
        'bundle.enterprise': environment.ENTERPRISE,
        'bundle.env': environment.NODE_ENV,

        ...SDK_INFO,
        ...BUILDKITE_INFO,
    })
}
