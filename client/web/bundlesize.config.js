const config = {
  files: [
    {
      // Our main app bundle, contains core logic that is loaded on every page.
      path: '../../ui/assets/scripts/app.*.js',
      maxSize: '200kb',
      compression: 'brotli',
    },
    // Our runtime bundle, loaded on every page and determines which other bundles are required
    {
      path: '../../ui/assets/scripts/runtime.*.js',
      maxSize: '10kb',
      compression: 'brotli',
    },
    // Our NPM dependencies
    {
      path: '../../ui/assets/scripts/npm.*.js',
      maxSize: '100kb',
      compression: 'brotli',
    },
    // Any exceptionally-large NPM dependencies
    {
      path: '../../ui/assets/scripts/npm-large.*.js',
      maxSize: '750kb',
      compression: 'brotli',
    },
  ],
}

module.exports = config
