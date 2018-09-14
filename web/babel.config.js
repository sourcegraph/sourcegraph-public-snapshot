// @ts-check

/** @type {import('@babel/core').TransformOptions} */
const config = {
  plugins: ['@babel/plugin-syntax-dynamic-import', 'babel-plugin-lodash'],
  presets: [
    [
      '@babel/preset-env',
      {
        useBuiltIns: 'entry',
        modules: false,
      },
    ],
  ],
}

module.exports = config
