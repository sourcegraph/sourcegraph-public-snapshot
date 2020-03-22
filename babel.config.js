// @ts-check

/** @type {import('@babel/core').ConfigFunction} */
module.exports = api => {
  const isTest = api.env('test')
  api.cache.forever()

  return {
    presets: [
      [
        '@babel/preset-env',
        {
          // Node (used for testing) doesn't support modules, so compile to CommonJS for testing.
          modules: isTest ? 'commonjs' : false,
          bugfixes: true,
          useBuiltIns: 'entry',
          corejs: 3,
        },
      ],
      '@babel/preset-typescript',
      '@babel/preset-react',
    ],
    plugins: [
      'babel-plugin-lodash',
      // Required to support typeoerm decorators in ./lsif
      ['@babel/plugin-proposal-decorators', { legacy: true }],
      // Node 12 (released 2019 Apr 23) supports these natively, but there seem to be issues when used with TypeScript.
      ['@babel/plugin-proposal-class-properties', { loose: true }],
    ],
  }
}
