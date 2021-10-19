// https://github.com/siddharthkp/bundlesize#configuration

module.exports = {
  files: [
    {
      path: './ui/assets/scripts/app.*.js',
      maxSize: '600kb',
      compression: 'brotli',
    },
  ],
}
