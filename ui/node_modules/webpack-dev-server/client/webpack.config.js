module.exports = {
	module: {
		loaders: [{
			test: /\.css$/,
			loader: "style!css"
		}, {
			test: /\.pug$/,
			loader: "pug?self"
		}]
	}
}
