import * as shelljs from "shelljs";

export function replaceWebpack(): void {
	const replaceTasks = [{
		from: "webpack/replace/JsonpMainTemplate.runtime.js",
		to: "node_modules/webpack/lib/JsonpMainTemplate.runtime.js",
	}, {
		from: "webpack/replace/log-apply-result.js",
		to: "node_modules/webpack/hot/log-apply-result.js",
	}];

	replaceTasks.forEach((task) => shelljs.cp(task.from, task.to));
};

export function copyAssets(type: string): void {
	const env = type === "build" ? "prod" : type;
	shelljs.rm("-rf", type);
	shelljs.mkdir(type);
	shelljs.cp(`chrome/manifest.${env}.json`, `${type}/manifest.json`);
	shelljs.cp("-R", "chrome/assets/", type);
	shelljs.exec(`jade -O "{ env: '${env}' }" -o ${type} chrome/views/`);
};
