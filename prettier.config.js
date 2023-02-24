const baseConfig = require('@sourcegraph/prettierrc')

module.exports = {
  ...baseConfig,
  plugins: [...(baseConfig.plugins || []), 'prettier-plugin-organize-imports'],
}
