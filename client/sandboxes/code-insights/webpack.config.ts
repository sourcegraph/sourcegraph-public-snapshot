
import path from 'path'

import ReactRefreshPlugin from '@pmmmwh/react-refresh-webpack-plugin';
import HtmlWebpackPlugin from 'html-webpack-plugin'

import { getCSSLoaders } from '@sourcegraph/web/dev/webpack/get-css-loaders'

const rootPath = path.resolve(__dirname, '..', '..', '..')

// eslint-disable-next-line import/no-default-export
export default {
  entry: path.resolve(__dirname, './src/demo.tsx'),
  output: {
    path: path.resolve(__dirname, 'dist'), // Note: Physical files are only output by the production build task `npm run build`.
    publicPath: '/',
    filename: '[name].bundle.js',
  },
  target: 'web', // necessary per https://webpack.github.io/docs/testing.html#compile-and-test
  mode: 'development',
  devtool: 'source-map',
  resolve: {
    alias: { react: require.resolve('react') },
    extensions: ['.ts', '.tsx', '.js', '.jsx'],
  },
  module: {
    rules: [
      {
        test: /\.(ts|js)x?$/,
        exclude: /node_modules/,
        use: ['babel-loader'],
      },
      {
        test: /\.(sass|scss)$/,
        // CSS Modules loaders are only applied when the file is explicitly named as CSS module stylesheet using the extension `.module.scss`.
        include: /\.module\.(sass|scss)$/,
        use: getCSSLoaders(rootPath, true, {
          loader: 'css-loader',
          options: {
            sourceMap: true,
            url: false,
            modules: {
              exportLocalsConvention: 'camelCase',
              localIdentName: '[name]__[local]_[hash:base64:5]',
            },
          },
        }),
      },
      {
        test: /\.(sass|scss)$/,
        exclude: /\.module\.(sass|scss)$/,
        use: getCSSLoaders(rootPath, true,{ loader: 'css-loader', options: { url: false } }),
      },
    ],
  },
  plugins: [
    new ReactRefreshPlugin({
      overlay: false,
    }),
    new HtmlWebpackPlugin({
      template: './src/index.html',
    }),
  ],
  devServer: {
    historyApiFallback: true,
  },
}
