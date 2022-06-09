import path from 'path'

import HtmlWebpackHarddiskPlugin from 'html-webpack-harddisk-plugin'
import HtmlWebpackPlugin, { TemplateParameter, Options } from 'html-webpack-plugin'
import signale from 'signale'
import { WebpackPluginInstance } from 'webpack'

import { createJsContext, ENVIRONMENT_CONFIG, STATIC_ASSETS_PATH } from '../utils'

const { SOURCEGRAPH_HTTPS_PORT, NODE_ENV } = ENVIRONMENT_CONFIG

export interface WebpackManifest {
    /** Main app entry JS bundle */
    'app.js': string
    /** Main app entry CSS bundle, only used in production mode */
    'app.css'?: string
    /** Runtime bundle, only used in development mode */
    'runtime.js'?: string
    /** React entry bundle, only used in production mode */
    'react.js'?: string
    /** If script files should be treated as JS modules. Required for esbuild bundle. */
    isModule?: boolean
}

/**
 * Returns an HTML page similar to `cmd/frontend/internal/app/ui/app.html` but when running
 * without the `frontend` service.
 *
 * Note: This page should be kept as close as possible to `app.html` to avoid any inconsistencies
 * between our development server and the actual production server.
 */
export const getHTMLPage = ({
    'app.js': appBundle,
    'app.css': cssBundle,
    'runtime.js': runtimeBundle,
    'react.js': reactBundle,
    isModule,
}: WebpackManifest): string => `
<!DOCTYPE html>
<html lang="en">
    <head>
        <title>Sourcegraph</title>
        <meta charset="utf-8">
        <meta name="viewport" content="width=device-width, viewport-fit=cover" />
        <meta name="referrer" content="origin-when-cross-origin"/>
        <meta name="color-scheme" content="light dark"/>
        ${cssBundle ? `<link rel="stylesheet" href="${cssBundle}">` : ''}
        ${
            ENVIRONMENT_CONFIG.SOURCEGRAPHDOTCOM_MODE
                ? '<script src="https://js.sentry-cdn.com/ae2f74442b154faf90b5ff0f7cd1c618.min.js" crossorigin="anonymous"></script>'
                : ''
        }
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

        ${runtimeBundle ? `<script src="${runtimeBundle}"></script>` : ''}
        ${reactBundle ? `<script src="${reactBundle}" ${isModule ? 'type="module"' : ''}></script>` : ''}
        <script src="${appBundle}" ${isModule ? 'type="module"' : ''}></script>
    </body>
</html>
`

/**
 * Search a list of file strings for a specific file.
 * Only uses the file prefix to allow matching against content-hashed filenames.
 */
const getBundleFromPath = (files: string[], filePrefix: string): string | undefined =>
    files.find(file => file.startsWith(`/.assets/${filePrefix}`))

export const getHTMLWebpackPlugins = (): WebpackPluginInstance[] => {
    signale.info('Serving `index.html` with `HTMLWebpackPlugin`.')

    const htmlWebpackPlugin = new HtmlWebpackPlugin({
        // `TemplateParameter` can be mutated. We need to tell TS that we didn't touch it.
        templateContent: (({ htmlWebpackPlugin }: TemplateParameter): string => {
            const { files } = htmlWebpackPlugin

            const appBundle = getBundleFromPath(files.js, 'scripts/app')

            if (!appBundle) {
                throw new Error('Could not find any entry bundle')
            }

            return getHTMLPage({
                'app.js': appBundle,
                'app.css': getBundleFromPath(files.css, 'styles/app'),
                'runtime.js': getBundleFromPath(files.js, 'scripts/runtime'),
                'react.js': getBundleFromPath(files.js, 'scripts/react'),
            })
        }) as Options['templateContent'],
        filename: path.resolve(STATIC_ASSETS_PATH, 'index.html'),
        alwaysWriteToDisk: true,
        inject: false,
    })

    // Write index.html to the disk so it can be served by dev/prod servers.
    return [htmlWebpackPlugin, new HtmlWebpackHarddiskPlugin()]
}
