import * as shelljs from "shelljs";
import * as tasks from "./tasks";

tasks.replaceWebpack();
console.info("[Copy assets]");
console.info("--------------------------------");
tasks.copyAssets("dev");

console.info("[Webpack Dev]");
console.info("--------------------------------");
shelljs.exec("webpack --config webpack/dev.config.js --progress --profile --colors --watch");
