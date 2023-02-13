const baseConfig = require('../../prettier.config.js')
module.exports = {
    ...baseConfig,
    plugins: [...(baseConfig.plugins || []), 'prettier-plugin-svelte'],
    pluginSearchDirs: [...(baseConfig.plugSearchDirs || []), './node_modules/'],
    overrides: [...(baseConfig.overrides || []), { files: '*.svelte', options: { parser: 'svelte' } }],
}
