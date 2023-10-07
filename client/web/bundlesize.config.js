const path = require('path')

function relativeAssets(base) {
  if (process.env.NODE_ENV !== undefined && process.env.NODE_ENV === 'development') {
    return path.join(base, '../../ui/assets')
  }
  return path.join(base, '../../ui/assets/enterprise')
}

const STATIC_ASSETS_PATH = process.env.WEB_BUNDLE_PATH || relativeAssets(__dirname)

const config = {
  files: [
    /**
     * Our main entry JavaScript bundles, contains core logic that is loaded on every page.
     */
    {
      path: path.join(STATIC_ASSETS_PATH, 'scripts/main*.js'),
      /**
       * Note: Temporary increase from 400kb.
       * Primary cause is due to multiple ongoing migrations that mean we are duplicating similar dependencies.
       * Issue to track: https://github.com/sourcegraph/sourcegraph/issues/37845
       */
      maxSize: '500kb',
    },
    {
      path: path.join(STATIC_ASSETS_PATH, 'scripts/embedMain*.js'),
      maxSize: '155kb',
    },
    {
      path: path.join(STATIC_ASSETS_PATH, 'chunks/chunk-*.js'),
      maxSize: '600kb', // 2 monaco chunks are very big
    },
    /**
     * Our main entry CSS bundle, contains core styles that are loaded on every page.
     */
    {
      path: path.join(STATIC_ASSETS_PATH, 'scripts/main*.css'),
      maxSize: '350kb',
    },
    /**
     * Notebook embed main entry CSS bundle.
     */
    {
      path: path.join(STATIC_ASSETS_PATH, 'scripts/embedMain*.css'),
      maxSize: '350kb',
    },
    {
      path: path.join(STATIC_ASSETS_PATH, 'chunks/chunk-*.css'),
      maxSize: '45kb',
    },
  ],
}

module.exports = config
