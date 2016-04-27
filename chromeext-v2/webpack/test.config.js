// for babel-plugin-webpack-loaders
const devConfig = require('./dev.config');

module.exports = {
	output: {
		libraryTarget: 'commonjs2'
	},
	module: {
		loaders: devConfig.module.loaders.slice(1)  // remove babel-loader
	}
};
