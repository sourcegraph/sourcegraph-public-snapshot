const path = require('path')

const ReactRefreshPlugin = require('@pmmmwh/react-refresh-webpack-plugin')
const HtmlWebpackPlugin = require('html-webpack-plugin')

module.exports = {
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
        test: /\.(scss)$/i,
        use: [
          'style-loader',
          {
            loader: 'css-loader',
            options: {
              url: false,
            },
          },
          'sass-loader',
        ],
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
}
