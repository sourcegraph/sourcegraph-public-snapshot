/**
 * The script collects web application bundlesize information from the disk and uploads it to Honeycomb.
 *
 * 1. Build web application using:
 * NODE_ENV=production DISABLE_TYPECHECKING=true pnpm build-web
 *
 * 2. Upload bundlesize information to Honeycomb:
 * HONEYCOMB_API_KEY=XXX pnpm --filter @sourcegraph/observability-server run bundlesize:web:upload
 *
 * 3. Check out collected data in Honeycomb! 🧠
 */
import { execSync } from 'child_process'
import path from 'path'

import { SemanticResourceAttributes } from '@opentelemetry/semantic-conventions'
import { cleanEnv, str } from 'envalid'

import { STATIC_ASSETS_PATH, WORKSPACES_PATH } from '@sourcegraph/build-config'

import { BUILDKITE_INFO, SDK_INFO } from '../constants'
import { libhoneySDK } from '../sdk'

import { getBundleSizeStats } from './getBundleSizeStats'

const environment = cleanEnv(process.env, {
    NODE_ENV: str({ choices: ['development', 'production'] }),
})

const bundleSizeStats = getBundleSizeStats({
    staticAssetsPath: STATIC_ASSETS_PATH,
    bundlesizeConfigPath: path.join(WORKSPACES_PATH, 'web/bundlesize.config'),
    webBuildManifestPath: path.join(STATIC_ASSETS_PATH, 'web.manifest.json'),
})

const commit = execSync('git rev-parse HEAD').toString().trim()
const branch = process.env.BUILDKITE_BRANCH || execSync('git rev-parse --abbrev-ref HEAD').toString().trim()

/**
 * Log every file size as a separate event to Honeycomb.
 */
for (const [baseFilePath, fileInfo] of Object.entries(bundleSizeStats)) {
    libhoneySDK.sendNow({
        name: 'bundlesize',
        [SemanticResourceAttributes.SERVICE_NAME]: 'bundlesize',
        [SemanticResourceAttributes.SERVICE_NAMESPACE]: 'web',
        [SemanticResourceAttributes.SERVICE_VERSION]: commit,
        'service.branch': branch,

        'bundle.file.name': baseFilePath,
        'bundle.file.size.raw': fileInfo.raw,
        'bundle.file.isInitial': fileInfo.isInitial,
        'bundle.file.isDynamicImport': fileInfo.isDynamicImport,
        'bundle.file.isCss': fileInfo.isCss,
        'bundle.file.isJs': fileInfo.isJs,
        'bundle.env': environment.NODE_ENV,

        ...SDK_INFO,
        ...BUILDKITE_INFO,
    })
}
