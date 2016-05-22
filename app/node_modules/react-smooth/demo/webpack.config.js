'use strict';

var path = require('path');

module.exports = {
  entry: {
    Gossip: __dirname + '/Gossip/app.js',
    GroupAnimation: __dirname + '/GroupAnimation/app.js',
    SimpleAnimation: __dirname + '/SimpleAnimation/app.js',
    Pendulum: __dirname + '/Pendulum/app.js',
  },
  devtool: 'inline-source-map',
  output: {
    path: __dirname,
    filename: '[name]/build.js',
  },
  resolve: {
    alias: {
      'react-smooth': path.join(__dirname, '..', 'src/index.js'),
    },
  },
  webpackServer: {
    hot: true,
  },
  module: {
    loaders: [
      {
        test: /\.js$/,
        loader: 'babel-loader',
        include: [
          __dirname,
          path.join(__dirname, '..', 'src'),
          path.join(__dirname, '..', 'node_modules'),
        ],
      },
    ],
  },
};
