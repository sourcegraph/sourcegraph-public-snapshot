const baseConfig = require('@sourcegraph/prettierrc')

module.exports = {
  ...baseConfig,
  plugins: [...(baseConfig.plugins || []), '@trivago/prettier-plugin-sort-imports'],
  importOrder: [
    '^react$',
    '<THIRD_PARTY_MODULES>',
    '^@sourcegraph/(.*)$',
    '^(?!.*.scss$)(../)+.*$', // Same dir paths, e.g. ""./Foo"
    '^(?!.*.scss$)(./)+.*$', // Higher dir paths, e.g. "../Foo", or "../../Foo"
    '.*.scss$', // SCSS imports
  ],
  importOrderSeparation: true,
}
