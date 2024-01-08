const MOCK_ASSETS_PATH = __dirname + '/assets'

module.exports = {
  files: [
    {
      path: MOCK_ASSETS_PATH + '/scripts/*.js',
      maxSize: '10kb',
    },
    {
      path: MOCK_ASSETS_PATH + '/styles/*.css',
      maxSize: '10kb',
    },
  ],
}
