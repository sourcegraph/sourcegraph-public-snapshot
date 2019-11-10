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
          useBuiltIns: 'entry',
          corejs: 3,
        },
      ],
      '@babel/preset-typescript',
      '@babel/preset-react',
    ],
    plugins: [
      '@babel/plugin-proposal-nullish-coalescing-operator',
      '@babel/plugin-proposal-optional-chaining',
      '@babel/plugin-syntax-dynamic-import',
      'babel-plugin-lodash',

      // Required to support typeoerm decorators in ./lsif
      ['@babel/plugin-proposal-decorators', { legacy: true }],
      // Node 12 (released 2019 Apr 23) supports these natively, so we can remove this plugin soon.
      ['@babel/plugin-proposal-class-properties', { loose: true }],
    ],
  }
}
