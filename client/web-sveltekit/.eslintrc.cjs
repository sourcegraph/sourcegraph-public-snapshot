const baseConfig = require('../../.eslintrc')
module.exports = {
    root: true,
    extends: ['../../.eslintrc.js', 'plugin:storybook/recommended'],
    parserOptions: {
        ...baseConfig.parserOptions,
        // This is set in the root config but doesn't work with Svelte components
        EXPERIMENTAL_useProjectService: false,
        project: [__dirname + '/tsconfig.json', __dirname + '/src/**/tsconfig.json'],
    },
    plugins: [...baseConfig.plugins, 'svelte3'],
    overrides: [
        ...baseConfig.overrides,
        {
            files: ['*.svelte'],
            processor: 'svelte3/svelte3',
        },
    ],
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
                cjs: 'always',
            },
        ],

        // These rules were newly introduced in @sourcegraph/eslint-config@0.35.0 and have not yet been
        // fixed in our existing code.
        'import/no-default-export': 'warn',
        'no-sparse-arrays': 'warn',
        '@typescript-eslint/explicit-function-return-type': 'warn',
        '@typescript-eslint/require-await': 'warn',
        'no-console': 'warn',
        '@typescript-eslint/ban-ts-comment': 'warn',
        '@typescript-eslint/no-floating-promises': 'warn',
        '@typescript-eslint/explicit-member-accessibility': 'warn',
    },
}
