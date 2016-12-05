module.exports = {
	module: {
		loaders: [{
			test: /\.css$/,
			loader: "style-loader!css-loader"
		}, {
			test: /\.pug$/,
			loader: "pug-loader?self"
		}]
	}
}
