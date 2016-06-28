var babelPresetEs2015,
    commonJsPlugin,
    es2015PluginList,
    es2015WebpackPluginList;

babelPresetEs2015 = require('babel-preset-es2015');

try {
    // npm ^2
    commonJsPlugin = require('babel-preset-es2015/node_modules/babel-plugin-transform-es2015-modules-commonjs');
} catch (error) {

}

if (!commonJsPlugin) {
    try {
        // npm ^3
        commonJsPlugin = require('babel-plugin-transform-es2015-modules-commonjs');
    } catch (error) {

    }
}

if (!commonJsPlugin) {
    throw new Error('Cannot resolve "babel-plugin-transform-es2015-modules-commonjs".');
}

es2015PluginList = babelPresetEs2015.plugins;

es2015WebpackPluginList = es2015PluginList.filter(function (es2015Plugin) {
    return es2015Plugin !== commonJsPlugin;
});

if (es2015PluginList.length !== es2015WebpackPluginList.length + 1) {
    throw new Error('Cannot remove "babel-plugin-transform-es2015-modules-commonjs" from the plugin list.');
}

module.exports = {
    plugins: es2015WebpackPluginList
};
