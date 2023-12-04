const baseConfig = require('../../prettier.config.js')
module.exports = {
    ...baseConfig,
    plugins: [...(baseConfig.plugins || []), 'prettier-plugin-svelte'],
    overrides: [...(baseConfig.overrides || []), { files: '*.svelte', options: { parser: 'svelte' } }],
}
