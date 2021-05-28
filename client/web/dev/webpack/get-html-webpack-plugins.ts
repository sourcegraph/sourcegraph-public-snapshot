import path from 'path'

import HtmlWebpackHarddiskPlugin from 'html-webpack-harddisk-plugin'
import HtmlWebpackPlugin, { TemplateParameter, Options } from 'html-webpack-plugin'
import { Plugin } from 'webpack'

import { createJsContext, environmentConfig, STATIC_ASSETS_PATH } from '../utils'

const { SOURCEGRAPH_HTTPS_PORT, NODE_ENV } = environmentConfig

export const getHTMLWebpackPlugins = (): Plugin[] => {
    const jsContext = createJsContext({ sourcegraphBaseUrl: `http://localhost:${SOURCEGRAPH_HTTPS_PORT}` })

    // TODO: use `cmd/frontend/internal/app/ui/app.html` template to be consistent with the default production setup.
    const templateContent = ({ htmlWebpackPlugin }: TemplateParameter): string => `
        <!DOCTYPE html>
        <html lang="en">
            <head>
                <meta charset="UTF-8">
                <title>Sourcegraph Development build</title>
                ${htmlWebpackPlugin.tags.headTags.toString()}
            </head>
            <body>
                <div id="root"></div>
                <script>
                    // Optional value useful for checking if index.html is created by HtmlWebpackPlugin with the right NODE_ENV.
                    window.webpackBuildEnvironment = '${NODE_ENV}'

                    // Required mock of the JS context object.
                    window.context = ${JSON.stringify(jsContext)}
                </script>
            </body>
        </html>
      `

    const htmlWebpackPlugin = new HtmlWebpackPlugin({
        // `TemplateParameter` can be mutated. We need to tell TS that we didn't touch it.
        templateContent: templateContent as Options['templateContent'],
        filename: path.resolve(STATIC_ASSETS_PATH, 'index.html'),
        alwaysWriteToDisk: true,
    })

    // Write index.html to the disk so it can be served by dev/prod servers.
    return [htmlWebpackPlugin, new HtmlWebpackHarddiskPlugin()]
}
