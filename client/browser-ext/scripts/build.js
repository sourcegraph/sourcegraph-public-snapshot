const fs = require("fs");
const hashFiles = require("hash-files")
const tasks = require("./tasks");

console.log("[Compute checksum]");
console.log("--------------------------------");
hashFiles({files: ["./app/**", "./chrome/**", "./scripts/**", "./webpack/**"]}, function(error, hash) {
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
    exec("webpack --config webpack/prod.config.js --progress --profile --colors");
});

