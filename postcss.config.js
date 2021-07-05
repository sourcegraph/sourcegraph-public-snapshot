module.exports = {
  plugins: [
    require('autoprefixer'),
    require('postcss-focus-visible'),
    // TODO(sqs): esbuild complains about @custom-media, so strip it...it's not supported yet by browsers anyway.
    require('postcss-custom-media')({ preserve: false }),
    require('postcss-inset'),
  ],
}
