const path = require('path')

const STATIC_ASSETS_PATH = path.join(__dirname, '../../ui/assets')

const config = {
  files: [
    /**
     * Our main entry JavaScript bundles, contains core logic that is loaded on every page.
     */
    {
      path: path.join(STATIC_ASSETS_PATH, 'scripts/app.bundle.js.br'),
      /**
       * Note: Temporary increase from 400kb.
       * Primary cause is due to multiple ongoing migrations that mean we are duplicating similar dependencies.
       * Issue to track: https://github.com/sourcegraph/sourcegraph/issues/37845
       */
      maxSize: '425kb',
      compression: 'none',
    },
    {
      path: path.join(STATIC_ASSETS_PATH, 'scripts/embed.bundle.js.br'),
      maxSize: '155kb',
      compression: 'none',
    },
    {
      path: path.join(STATIC_ASSETS_PATH, 'scripts/react.bundle.js.br'),
      maxSize: '45kb',
      compression: 'none',
    },
    {
      path: path.join(STATIC_ASSETS_PATH, 'scripts/opentelemetry.bundle.js.br'),
      maxSize: '30kb',
      compression: 'none',
    },
    /**
     * Our generated application chunks. Matches the deterministic id generated by Webpack.
     *
     * Note: The vast majority of our chunks are under 200kb, this threshold is bloated as we treat the Monaco editor as a normal chunk.
     * We should consider not doing this, as it is much larger than other chunks and we would likely benefit from caching this differently.
     * Issue to improve this: https://github.com/sourcegraph/sourcegraph/issues/26573
     */
    {
      path: path.join(STATIC_ASSETS_PATH, 'scripts/!(sg_)*.chunk.js.br'),
      maxSize: '500kb',
      compression: 'none',
    },
    {
      path: path.join(STATIC_ASSETS_PATH, 'scripts/sg_*.chunk.js.br'),
      maxSize: '200kb',
      compression: 'none',
    },
    /**
     * Our generated worker files.
     */
    {
      path: path.join(STATIC_ASSETS_PATH, '*.worker.js.br'),
      maxSize: '250kb',
      compression: 'none',
    },
    /**
     * Our main entry CSS bundle, contains core styles that are loaded on every page.
     */
    {
      path: path.join(STATIC_ASSETS_PATH, 'styles/app.*.css.br'),
      maxSize: '50kb',
      compression: 'none',
    },
    /**
     * Notebook embed main entry CSS bundle.
     */
    {
      path: path.join(STATIC_ASSETS_PATH, 'styles/embed.*.css.br'),
      maxSize: '35kb',
      compression: 'none',
    },
    {
      path: path.join(STATIC_ASSETS_PATH, 'styles/!(app|embed).*.css.br'),
      maxSize: '10kb',
      compression: 'none',
    },
  ],
}

module.exports = config
