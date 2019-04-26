// @ts-check

/** @type {import('@babel/core').TransformOptions} */
const config = {
  plugins: ['@babel/plugin-syntax-dynamic-import', 'babel-plugin-lodash'],
  presets: [
    [
      '@babel/preset-env',
      {
        modules: false,
        targets: require('../package.json').browserslist,
        useBuiltIns: 'entry',
      },
    ],
  ],
}

module.exports = config
