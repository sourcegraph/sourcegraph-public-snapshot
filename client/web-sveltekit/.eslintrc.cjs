const baseConfig = require('../../.eslintrc')
module.exports = {
    extends: '../../.eslintrc.js',
    parserOptions: {
        ...baseConfig.parserOptions,
        project: [__dirname + '/tsconfig.json', __dirname + '/src/**/tsconfig.json'],
    },
    plugins: [...baseConfig.plugins, 'svelte3'],
    overrides: [...baseConfig.overrides, { files: ['*.svelte'], processor: 'svelte3/svelte3' }],
    settings: {
        ...baseConfig.settings,
        'svelte3/typescript': () => require('typescript'),
    },
    rules: {
        ...baseConfig.rules,
        'import/extensions': [
            'error',
            'never',
            {
                svelte: 'always',
                svg: 'always',
            },
        ],
    },
}
