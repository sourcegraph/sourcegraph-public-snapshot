const tasks = require('./tasks');

tasks.replaceWebpack();
console.log('[Copy assets]');
console.log('--------------------------------');
tasks.copyAssets('dev');

console.log('[Webpack Dev]');
console.log('--------------------------------');
console.log('Please allow `https://localhost:3000` connections in Google Chrome');
console.log('and load unpacked extensions with `./dev` folder.  (see https://developer.chrome.com/extensions/getstarted#unpacked)\n');
exec('webpack-dev-server --config=webpack/dev.config.js --no-info --hot --inline --colors');
