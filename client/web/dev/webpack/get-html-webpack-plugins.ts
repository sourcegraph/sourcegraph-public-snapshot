import path from 'path'

import HtmlWebpackHarddiskPlugin from 'html-webpack-harddisk-plugin'
import HtmlWebpackPlugin, { TemplateParameter, Options } from 'html-webpack-plugin'
import { Plugin } from 'webpack'

import { createJsContext, environmentConfig, STATIC_ASSETS_PATH } from '../utils'

const { SOURCEGRAPH_HTTPS_PORT, SOURCEGRAPH_API_URL, NODE_ENV } = environmentConfig

export const getHTMLWebpackPlugins = (): Plugin[] => {
    const jsContext = createJsContext({ sourcegraphBaseUrl: `http://localhost:${SOURCEGRAPH_HTTPS_PORT}` })

    // TODO: use `cmd/frontend/internal/app/ui/app.html` template to be consistent with default production setup.
    const templateContent = ({ htmlWebpackPlugin }: TemplateParameter): string => `
        <html>
            <head>
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

                    // On https://k8s.sgdev.org unauthorized user receives 401 error and sees error page.
                    // This helper is added to redirect to sign-in page automatically.
                    window.addEventListener('error', function(event) {
                      const signInRoute = '/sign-in'

                      if (
                          '${SOURCEGRAPH_API_URL}'.includes('k8s.sgdev.org') &&
                          event.message.includes('401 Unauthorized') &&
                           window.location.pathname !== signInRoute
                        ) {
                        window.location.href = signInRoute;
                      }
                    })
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
