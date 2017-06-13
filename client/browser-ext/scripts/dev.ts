// tslint:disable

import * as tasks from "./tasks";
import * as shelljs from "shelljs";

tasks.replaceWebpack();
console.log("[Copy assets]");
console.log("--------------------------------");
tasks.copyAssets("dev");

console.log("[Webpack Dev]");
console.log("--------------------------------");
shelljs.exec("webpack --config webpack/dev.config.js --progress --profile --colors --watch");
