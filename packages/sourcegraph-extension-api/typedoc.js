module.exports = {
  name: 'Sourcegraph extension API',
  out: 'dist/docs/',
  readme: 'none',
  includes: './src',
  exclude: [
    '**/client/**/*',
    '**/common/**/*',
    '**/extension/**/*',
    '**/integration-test/**/*',
    '**/protocol/**/*',
    '**/util*.ts',
  ],
  includeDeclarations: true,
  mode: 'file',
  excludeExternals: true,
  excludePrivate: true,
}
