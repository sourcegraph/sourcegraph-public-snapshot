const path = require('path')
const { remove } = require('lodash')
const { DefinePlugin, ProgressPlugin } = require('webpack')
const MonacoWebpackPlugin = require('monaco-editor-webpack-plugin')

const monacoEditorPaths = [path.resolve(__dirname, '..', 'node_modules', 'monaco-editor')]

const config = {
  stories: ['../**/*.story.tsx'],
  addons: ['@storybook/addon-knobs', '@storybook/addon-actions', '@storybook/addon-options', 'storybook-addon-designs'],
  /**
   * @param config {import('webpack').Configuration}
   * @returns {import('webpack').Configuration}
   */
  webpackFinal: config => {
    // Include sourcemaps
    config.mode = 'development'
    const definePlugin = config.plugins.find(plugin => plugin instanceof DefinePlugin)
    // @ts-ignore
    definePlugin.definitions.NODE_ENV = JSON.stringify('development')
    // @ts-ignore
    definePlugin.definitions['process.env'].NODE_ENV = JSON.stringify('development')

    // We don't use Storybook's default config for our repo, it doesn't handle TypeScript.
    config.module.rules.splice(0, 1)

    if (process.env.CI) {
      remove(config.plugins, plugin => plugin instanceof ProgressPlugin)
    }

    config.module.rules.push({
      test: /\.tsx?$/,
      loader: require.resolve('babel-loader'),
      options: {
        configFile: path.resolve(__dirname, '..', 'babel.config.js'),
      },
    })

    config.resolve.extensions.push('.ts', '.tsx')

    config.plugins.push(
      new MonacoWebpackPlugin({
        languages: ['json'],
        features: [
          'bracketMatching',
          'clipboard',
          'coreCommands',
          'cursorUndo',
          'find',
          'format',
          'hover',
          'inPlaceReplace',
          'iPadShowKeyboard',
          'links',
          'suggest',
        ],
      })
    )

    const storybookDirectory = path.resolve(__dirname, '../node_modules/@storybook')

    // Put our style rules at the beginning so they're processed by the time it
    // gets to storybook's style rules.
    config.module.rules.unshift({
      test: /\.(sass|scss)$/,
      use: [
        'to-string-loader',
        'css-loader',
        {
          loader: 'sass-loader',
          options: {
            sassOptions: {
              includePaths: [path.resolve(__dirname, '..', 'node_modules')],
            },
          },
        },
      ],
      // Make sure Storybook styles get handled by the Storybook config
      exclude: storybookDirectory,
    })

    config.module.rules.unshift({
      // CSS rule for monaco-editor and other external plain CSS (skip SASS and PostCSS for build perf)
      test: /\.css$/,
      include: monacoEditorPaths,
      // Make sure Storybook styles get handled by the Storybook config
      exclude: storybookDirectory,
      use: ['to-string-loader', 'css-loader'],
    })

    Object.assign(config.entry, {
      'editor.worker': 'monaco-editor/esm/vs/editor/editor.worker.js',
      'json.worker': 'monaco-editor/esm/vs/language/json/json.worker',
    })

    // Make sure Storybook style loaders are only evaluated for Storybook styles.
    config.module.rules.find(rule => rule.test?.toString() === /\.css$/.toString()).include = storybookDirectory

    return config
  },
}
module.exports = config
