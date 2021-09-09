import path from 'path'

import HtmlWebpackHarddiskPlugin from 'html-webpack-harddisk-plugin'
import HtmlWebpackPlugin, { TemplateParameter, Options } from 'html-webpack-plugin'
import { WebpackPluginInstance } from 'webpack'

import { createJsContext, environmentConfig, STATIC_ASSETS_PATH } from '../utils'

const { SOURCEGRAPH_HTTPS_PORT, NODE_ENV } = environmentConfig

interface HTMLPageData {
    head: string
    bodyEnd: string
}

/**
 * Returns an HTML page similar to `cmd/frontend/internal/app/ui/app.html` but when running
 * without the `frontend` service.
 */
export const getHTMLPage = ({ head, bodyEnd }: HTMLPageData): string => `
<!DOCTYPE html>
<html lang="en">
    <head>
        <title>Sourcegraph</title>
        <meta charset="utf-8">
        <meta name="viewport" content="width=device-width, viewport-fit=cover" />
        <meta name="referrer" content="origin-when-cross-origin"/>
        <meta name="color-scheme" content="light dark"/>
        ${head}
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
        ${bodyEnd}
    </body>
</html>
`

export const getHTMLWebpackPlugins = (): WebpackPluginInstance[] => {
    const htmlWebpackPlugin = new HtmlWebpackPlugin({
        // `TemplateParameter` can be mutated. We need to tell TS that we didn't touch it.
        templateContent: (({ htmlWebpackPlugin }: TemplateParameter): string =>
            getHTMLPage({
                head: htmlWebpackPlugin.tags.headTags.filter(tag => tag.tagName !== 'script').toString(),
                bodyEnd: htmlWebpackPlugin.tags.headTags.filter(tag => tag.tagName === 'script').toString(),
            })) as Options['templateContent'],
        filename: path.resolve(STATIC_ASSETS_PATH, 'index.html'),
        alwaysWriteToDisk: true,
        inject: false,
    })

    // Write index.html to the disk so it can be served by dev/prod servers.
    return [htmlWebpackPlugin, new HtmlWebpackHarddiskPlugin()]
}
