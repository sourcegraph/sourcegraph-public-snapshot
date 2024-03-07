const path = require('path')

const STATIC_ASSETS_PATH = path.join(process.env.WEB_BUNDLE_PATH || path.join(__dirname, 'dist'), 'assets')

const config = {
  files: [
    /**
     * Our main entry JavaScript bundles, contains core logic that is loaded on every page.
     */
    {
      path: path.join(STATIC_ASSETS_PATH, 'main*.js'),
      /**
       * Note: Temporary increase from 400kb.
       * Primary cause is due to multiple ongoing migrations that mean we are duplicating similar dependencies.
       * Issue to track: https://github.com/sourcegraph/sourcegraph/issues/37845
       */
      maxSize: '1000kb',
    },
    // {
    //   path: path.join(STATIC_ASSETS_PATH, 'embedMain*.js'),
    //   maxSize: '155kb',
    // },
    {
      path: path.join(STATIC_ASSETS_PATH, '*.js'),
      maxSize: '600kb', // 2 monaco chunks are very big
    },
    /**
     * Our main entry CSS bundle, contains core styles that are loaded on every page.
     */
    {
      path: path.join(STATIC_ASSETS_PATH, 'style*.css'),
      maxSize: '350kb',
    },
    /**
     * Notebook embed main entry CSS bundle.
     */
    // {
    //   path: path.join(STATIC_ASSETS_PATH, 'embedMain*.css'),
    //   maxSize: '350kb',
    // },
    {
      path: path.join(STATIC_ASSETS_PATH, '*.css'),
      maxSize: '45kb',
    },
  ],
}

module.exports = config
