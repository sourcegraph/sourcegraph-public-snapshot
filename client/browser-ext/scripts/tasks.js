require("shelljs/global");

exports.replaceWebpack = function() {
	const replaceTasks = [{
		from: "webpack/replace/JsonpMainTemplate.runtime.js",
		to: "node_modules/webpack/lib/JsonpMainTemplate.runtime.js",
	}, {
		from: "webpack/replace/log-apply-result.js",
		to: "node_modules/webpack/hot/log-apply-result.js",
	}];

	replaceTasks.forEach((task) => cp(task.from, task.to));
};

exports.copyAssets = function(type) {
	const env = type === "build" ? "prod" : type;
	rm("-rf", type);
	mkdir(type);
	cp(`chrome/manifest.${env}.json`, `${type}/manifest.json`);
	cp("-R", "chrome/assets/", type);
	cp("Dockerfile.selenium", `${type}/Dockerfile`)
	exec(`jade -O "{ env: '${env}' }" -o ${type} chrome/views/`);
};

