module.exports = {
  plugins: [
    require('autoprefixer'),
    require('postcss-focus-visible'),
    require('postcss-pxtorem')({ propWhiteList: [] }),
  ],
}
