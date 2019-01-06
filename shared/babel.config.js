// @ts-check

/** @type {import('@babel/core').TransformOptions} */
const config = {
  plugins: ['@babel/plugin-syntax-dynamic-import'],
  presets: ['@babel/preset-env'],
}

module.exports = config
