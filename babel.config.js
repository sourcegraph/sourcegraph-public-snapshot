// @ts-check

/** @type {import('@babel/core').TransformOptions} */
const config = {
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
