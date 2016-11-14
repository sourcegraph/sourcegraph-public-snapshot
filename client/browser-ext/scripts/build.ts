// tslint:disable

import * as tasks from "./tasks";
import * as fs from "fs";
import * as shelljs from "shelljs";

console.log("[Compute checksum]");
console.log("--------------------------------");
require("hash-files")({files: ["./app/**", "./chrome/**", "./scripts/**", "./webpack/**"]}, (error, hash) => {
	if (error) {
		console.error(error);
		return;
	}

	try {
		const savedHash = fs.readFileSync(".checksum", "utf8");
		if (savedHash === hash) {
			console.log("Match checksum, skipping build...");
			return;
		}
	} catch (e) {}

	fs.writeFileSync(".checksum", hash)

	tasks.replaceWebpack();
	console.log("[Copy assets]");
	console.log("--------------------------------");
	tasks.copyAssets("build");

	console.log("[Webpack Build]");
	console.log("--------------------------------");
	shelljs.exec("webpack --config webpack/prod.config.js --progress --profile --colors");
});

