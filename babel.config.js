// @ts-check

/** @type {import('@babel/core').ConfigFunction} */
module.exports = api => {
  const isTest = api.env('test')
  api.cache.forever()

  /**
   * Do no use babel-preset-env for mocha tests transpilation in Bazel.
   * This is temporary workaround to allow us to use modern language featurs in `drive.page.evaluate` calls.
   */
  const disablePresetEnv = Boolean(process.env.DISABLE_PRESET_ENV && JSON.parse(process.env.DISABLE_PRESET_ENV))

  return {
    presets: [
      ...(disablePresetEnv
        ? []
        : [
            [
              '@babel/preset-env',
              {
                // Node (used for testing) doesn't support modules, so compile to CommonJS for testing.
                modules: process.env.BABEL_MODULE ?? (isTest ? 'commonjs' : false),
              },
            ],
          ]),
      ['@babel/preset-typescript', { isTSX: true, allExtensions: true }],
      [
        '@babel/preset-react',
        {
          runtime: 'automatic',
        },
      ],
    ],
  }
}
