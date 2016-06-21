// for babel-plugin-webpack-loaders
const config = require("./prod.config");

module.exports = {
	output: {
		libraryTarget: "commonjs2"
	},
	module: {
		loaders: config.module.loaders.slice(1)  // remove babel-loader
	}
};
