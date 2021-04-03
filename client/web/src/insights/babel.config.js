module.exports = {
  'plugins': [
    ['@babel/plugin-proposal-class-properties', { 'loose': true }],
    '@babel/plugin-proposal-optional-chaining',
    '@babel/plugin-proposal-nullish-coalescing-operator',
    'react-refresh/babel',
  ],
  'presets': [
    ['@babel/preset-env', { targets: { node: true } } ],
    '@babel/preset-typescript',
    '@babel/preset-react',
  ],
};
