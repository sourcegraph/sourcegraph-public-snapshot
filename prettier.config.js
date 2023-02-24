const baseConfig = require('@sourcegraph/prettierrc')

module.exports = {
  ...baseConfig,
  plugins: [...(baseConfig.plugins || []), '@trivago/prettier-plugin-sort-imports'],
  importOrder: ['^react$', '<THIRD_PARTY_MODULES>', '^@sourcegraph/(.*)$', '^(?!.*.scss$)[./].*$', '^.*.scss$'],
  importOrderSeparation: true,
  importOrderSortSpecifiers: false,
}
