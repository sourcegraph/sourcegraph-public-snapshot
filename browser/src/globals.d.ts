interface Window {
    SOURCEGRAPH_URL: string | undefined
    PHABRICATOR_CALLSIGN_MAPPINGS:
        | {
              callsign: string
              path: string
          }[]
        | undefined
    SOURCEGRAPH_PHABRICATOR_EXTENSION: boolean | undefined
    SG_ENV: 'EXTENSION' | 'PAGE'
    EXTENSION_ENV: 'CONTENT' | 'BACKGROUND' | 'OPTIONS' | null
    SOURCEGRAPH_BUNDLE_URL: string | undefined // Bundle Sourcegraph URL is set from the Phabricator extension.
    // Bitbucket has a global require function on the DOM that we rely on to get the current Bitbucket state.
    require: any
}

/**
 * Set by shared/dev/jest-environment.js
 */
declare var jsdom: import('jsdom').JSDOM

/**
 * For Web Worker entrypoints using Webpack's worker-loader.
 *
 * See https://github.com/webpack-contrib/worker-loader#integrating-with-typescript.
 */
declare module 'worker-loader?inline*' {
    class WebpackWorker extends Worker {
        constructor()
    }
    export default WebpackWorker
}
