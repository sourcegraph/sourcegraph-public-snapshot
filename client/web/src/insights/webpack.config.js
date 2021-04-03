const path = require('path');
const webpack = require('webpack');
const HtmlWebpackPlugin = require('html-webpack-plugin');
const ReactRefreshPlugin = require('@pmmmwh/react-refresh-webpack-plugin');
const ForkTsCheckerWebpackPlugin = require('fork-ts-checker-webpack-plugin');

module.exports = {
  entry: path.resolve(__dirname, './web-demo/demo.tsx'),
  output: {
    path: path.resolve(__dirname, 'dist'), // Note: Physical files are only output by the production build task `npm run build`.
    publicPath: '/',
    filename: '[name].bundle.js'
  },
  target: 'web', // necessary per https://webpack.github.io/docs/testing.html#compile-and-test
  mode: 'development',
  devtool: 'source-map',
  resolve: {
    alias: { react: require.resolve('react') },
    extensions: ['.ts', '.tsx', '.js', '.jsx',]
  },
  module: {
    rules: [
      {
        test: /\.(ts|js)x?$/,
        exclude: /node_modules\/(?!(@wrike-kit)\/).*/,
        use: ['babel-loader'],
      },
      {
        test: /\.(png|jpe?g|gif)$/i,
        use: [
          {
            loader: 'file-loader',
            options: { name: 'static/media/[name].[hash:8].[ext]' }
          },
        ],
      },
      {
        test: /\.(scss)$/i,
        use: [
          'style-loader',
          {
            loader: 'css-loader',
          },
          'sass-loader'
        ],
      },
      {
        test: /\.css$/i,
        use: [
          'style-loader',
          {
            loader: 'css-loader',
          },
          'postcss-loader'
        ],
      },
    ]
  },
  plugins: [
    new ReactRefreshPlugin({
      overlay: false
    }),
    new HtmlWebpackPlugin({
      template: './web-demo/index.html'
    }),
    // new ForkTsCheckerWebpackPlugin(),
  ]
};
