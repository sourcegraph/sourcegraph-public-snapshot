import * as fs from "fs";
import * as shelljs from "shelljs";
import * as tasks from "./tasks";

console.info("[Compute checksum]");
console.info("--------------------------------");
require("hash-files")({ files: ["./app/**", "./chrome/**", "./phabricator/**", "./scripts/**", "./webpack/**"] }, (error, hash) => {
	if (error) {
		console.error(error);
		return;
	}

	try {
		const savedHash = fs.readFileSync(".checksum", "utf8");
		if (savedHash === hash) {
			console.info("Match checksum, skipping build...");
			return;
		}
	} catch (e) {
		// ignore
	}

	fs.writeFileSync(".checksum", hash);

	tasks.replaceWebpack();
	console.info("[Copy assets]");
	console.info("--------------------------------");
	tasks.copyAssets("build");

	console.info("[Webpack Build]");
	console.info("--------------------------------");
	shelljs.exec("webpack --config webpack/prod.config.js --progress --profile --colors");
});
