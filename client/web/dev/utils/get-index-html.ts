import { readFileSync } from 'fs'
import path from 'path'

import type { SourcegraphContext } from '../../src/jscontext'

import { WEB_BUILD_MANIFEST_FILENAME, assetPathPrefix } from './constants'
import { createJsContext } from './create-js-context'
import { ENVIRONMENT_CONFIG, HTTPS_WEB_SERVER_URL } from './environment-config'
import type { WebBuildManifest } from './webBuildManifest'

const { STATIC_ASSETS_PATH } = ENVIRONMENT_CONFIG

const WEB_BUILD_MANIFEST_PATH = path.resolve(STATIC_ASSETS_PATH, WEB_BUILD_MANIFEST_FILENAME)
export const HTML_INDEX_PATH = path.resolve(STATIC_ASSETS_PATH, 'index.html')

export const getWebBuildManifest = (): WebBuildManifest =>
    JSON.parse(readFileSync(WEB_BUILD_MANIFEST_PATH, 'utf-8')) as WebBuildManifest

interface GetHTMLPageOptions {
    manifest: WebBuildManifest
    /**
     * Used to inject dummy `window.context` in integration tests.
     */
    jsContext?: SourcegraphContext
    /**
     * Used to inject `window.context` received from the API proxy.
     */
    jsContextScript?: string
}

/**
 * Returns an HTML page similar to `cmd/frontend/internal/app/ui/app.html` but when running
 * without the `frontend` service.
 *
 * Note: This page should be kept as close as possible to `app.html` to avoid any inconsistencies
 * between our development server and the actual production server.
 */
export function getIndexHTML({ manifest, jsContext, jsContextScript }: GetHTMLPageOptions): string {
    if (!manifest.assets['src/enterprise/main']) {
        throw new Error('entrypoint asset not found')
    }
    return `
<!DOCTYPE html>
<html lang="en">
    <head>
        <title>Sourcegraph</title>
        <meta charset="utf-8">
        <meta name="viewport" content="width=device-width, viewport-fit=cover" />
        <meta name="referrer" content="origin-when-cross-origin"/>
        <meta name="color-scheme" content="light dark"/>
        ${
            ENVIRONMENT_CONFIG.SOURCEGRAPHDOTCOM_MODE
                ? '<script src="https://js.sentry-cdn.com/ae2f74442b154faf90b5ff0f7cd1c618.min.js" crossorigin="anonymous"></script>'
                : ''
        }
        ${
            manifest.assets['src/enterprise/main']?.css
                ? `<link rel="stylesheet" href="${assetPathPrefix}/${manifest.assets['src/enterprise/main']?.css}">`
                : ''
        }
    </head>
    <body>
        ${manifest.devInjectHTML ?? ''}
        <div id="root"></div>
        <script>
            ${
                jsContextScript ||
                `
                // Required mock of the JS context object.
                window.context = ${JSON.stringify(
                    jsContext ?? createJsContext({ sourcegraphBaseUrl: `${HTTPS_WEB_SERVER_URL}` })
                )}
            `
            }
        </script>

        <script src="${assetPathPrefix}/${manifest.assets['src/enterprise/main'].js}" type="module"></script>
    </body>
</html>
`
}
