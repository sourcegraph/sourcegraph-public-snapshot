import baseConfig from '../../prettier.config.js'

export default {
  ...baseConfig,
  plugins: [...(baseConfig.plugins || []), 'prettier-plugin-svelte'],
  overrides: [...(baseConfig.overrides || []), { files: '*.svelte', options: { parser: 'svelte' } }],
}
