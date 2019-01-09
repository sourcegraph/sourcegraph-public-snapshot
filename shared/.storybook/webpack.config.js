// @ts-check

const path = require('path')

module.exports = (baseConfig, env, config) => {
  config.module.rules.push({
    test: /\.(ts|tsx)$/,
    loader: require.resolve('babel-loader'),
    options: {
      presets: [['react-app', { flow: false, typescript: true }]],
    },
  })
  config.resolve.extensions.push('.ts', '.tsx')

  // Put our style rules at the beginning so they're processed by the time it
  // gets to storybook's style rules.
  config.module.rules.unshift({
    test: /\.(css|sass|scss)$/,
    use: [
      'style-loader',
      'postcss-loader',
      {
        loader: 'sass-loader',
        options: {
          includePaths: [path.resolve(__dirname, '../..', 'node_modules')],
        },
      },
    ],
  })

  return config
}
