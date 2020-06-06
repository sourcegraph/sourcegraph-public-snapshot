// @ts-check

const config = require('../jest.config.base')

const exportedConfig = {
  ...config,
  displayName: 'storybooks',
  rootDir: __dirname,
  collectCoverage: false, // Collected through Puppeteer
  roots: ['<rootDir>'],
  verbose: true,
  transform: {
    '^.+\\.[tj]sx?$': 'babel-jest',
    '^.+\\.mdx$': '@storybook/addon-docs/jest-transform-mdx',
  },
}

module.exports = exportedConfig
