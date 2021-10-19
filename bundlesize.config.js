// https://github.com/siddharthkp/bundlesize#configuration

module.exports = {
  files: [
    {
      path: './ui/assets/scripts/*.js',
      maxSize: '1000kb',
      compression: 'brotli',
    },
  ],
}
