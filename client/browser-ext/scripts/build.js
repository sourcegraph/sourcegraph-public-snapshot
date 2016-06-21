const tasks = require("./tasks");

tasks.replaceWebpack();
console.log("[Copy assets]");
console.log("--------------------------------");
tasks.copyAssets("build");

console.log("[Webpack Build]");
console.log("--------------------------------");
exec("webpack --config webpack/prod.config.js --progress --profile --colors");
