// @ts-check

/** @type {import('@babel/core').ConfigFunction} */
module.exports = {
  presets: [['@babel/preset-env', { targets: { node: 'current' } }], '@babel/preset-typescript'],
}
