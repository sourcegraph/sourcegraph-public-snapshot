import path from 'path'

import HtmlWebpackHarddiskPlugin from 'html-webpack-harddisk-plugin'
import HtmlWebpackPlugin from 'html-webpack-plugin'
import { WebpackPluginInstance } from 'webpack'

import { createJsContext, environmentConfig, STATIC_ASSETS_PATH } from '../utils'

import { getWebpackManifest, WebpackManifest } from './get-manifest'

const { SOURCEGRAPH_HTTPS_PORT, NODE_ENV } = environmentConfig

interface GetHTMLPageParameters {
    manifest: WebpackManifest
}

/**
 * Returns an HTML page similar to `cmd/frontend/internal/app/ui/app.html` but when running
 * without the `frontend` service.
 *
 * Note: This page should be kept as close as possible to `app.html`, to avoid any inconsistencies
 * between our development server and the actual production server.
 */
export const getHTMLPage = ({ manifest }: GetHTMLPageParameters): string => {
    const {
        'app.js': appBundlePath,
        'react.js': reactBundlePath,
        'runtime.js': runtimeBundlePath,
        'app.css': cssBundlePath,
        isModule,
    } = manifest

    return `
<!DOCTYPE html>
<html lang="en">
    <head>
        <title>Sourcegraph</title>
        <meta charset="utf-8">
        <meta name="viewport" content="width=device-width, viewport-fit=cover" />
        <meta name="referrer" content="origin-when-cross-origin"/>
        <meta name="color-scheme" content="light dark"/>
        ${cssBundlePath ? `<link rel="stylesheet" href="${cssBundlePath}">` : ''}
    </head>
    <body>
        <div id="root"></div>
        <script>
            // Optional value useful for checking if index.html is created by HtmlWebpackPlugin with the right NODE_ENV.
            window.webpackBuildEnvironment = '${NODE_ENV}'

            // Required mock of the JS context object.
            window.context = ${JSON.stringify(
                createJsContext({ sourcegraphBaseUrl: `http://localhost:${SOURCEGRAPH_HTTPS_PORT}` })
            )}
        </script>

        ${runtimeBundlePath ? `<script src="${runtimeBundlePath}"></script>` : ''}
        ${reactBundlePath ? `<script src="${reactBundlePath}" ${isModule ? 'type="module"' : ''}></script>` : ''}
        <script src="${appBundlePath}" ${isModule ? 'type="module"' : ''}></script>
    </body>
</html>
`
}

export const getHTMLWebpackPlugins = (): WebpackPluginInstance[] => {
    const htmlWebpackPlugin = new HtmlWebpackPlugin({
        templateContent: getHTMLPage({ manifest: getWebpackManifest() }),
        filename: path.resolve(STATIC_ASSETS_PATH, 'index.html'),
        alwaysWriteToDisk: true,
        inject: false,
    })

    // Write index.html to the disk so it can be served by dev/prod servers.
    return [htmlWebpackPlugin, new HtmlWebpackHarddiskPlugin()]
}
