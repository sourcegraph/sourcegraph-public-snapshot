const path = require('path')
const sassImportOnce = require('node-sass-import-once')

module.exports = (baseConfig, env, config) => {
  config.module.rules.push({
    test: /\.tsx?$/,
    use: [
      {
        loader: 'babel-loader',
        options: {
          cacheDirectory: true,
        },
      },
      {
        loader: 'ts-loader',
        options: {
          compilerOptions: {
            target: 'es6',
            module: 'esnext',
            noEmit: false, // tsconfig.json sets this to true to avoid output when running tsc manually
          },
          transpileOnly: process.env.DISABLE_TYPECHECKING === 'true',
        },
      },
    ],
  })

  config.module.rules.push({
    test: /\.jsx?$/,
    loader: {
      loader: 'babel-loader',
      options: {
        cacheDirectory: true,
      },
    },
  })

  config.module.rules.unshift({
    test: /\.(css|sass|scss)$/,
    use: [
      'style-loader',
      {
        loader: 'css-loader',
        options: {
          minimize: process.env.NODE_ENV === 'production',
        },
      },
      'postcss-loader',
      {
        loader: 'sass-loader',
        options: {
          includePaths: [path.resolve(__dirname, '..', 'node_modules')],
          importer: sassImportOnce,
          importOnce: {
            css: true,
          },
        },
      },
    ],
  })

  config.resolve.extensions.push('.ts', '.tsx')
  return config
}
