// @ts-check

const config = {
  extends: '@sourcegraph/eslint-config',
  env: {
    browser: true,
    node: true,
    es6: true,
  },
  parserOptions: {
    ecmaVersion: 2018,
    sourceType: 'module',
    ecmaFeatures: {
      jsx: true,
    },
    EXPERIMENTAL_useSourceOfProjectReferenceRedirect: true,
    project: __dirname + '/tsconfig.json',
  },
  settings: {
    react: {
      version: 'detect',
    },
    linkComponents: [
      {
        name: 'LinkOrSpan',
        linkAttribute: 'to',
      },
      {
        name: 'Link',
        linkAttribute: 'to',
      },
    ],
  },
  plugins: ['@sourcegraph/sourcegraph', 'monorepo', '@sourcegraph/wildcard'],
  rules: {
    // Rules that are specific to this repo
    // All other rules should go into https://github.com/sourcegraph/eslint-config
    'monorepo/no-relative-import': 'error',
    '@sourcegraph/sourcegraph/check-help-links': 'error',
    'no-restricted-imports': [
      'error',
      {
        paths: [
          'highlight.js',
          'marked',
          'rxjs/ajax',
          {
            name: 'rxjs',
            importNames: ['animationFrameScheduler'],
            message: 'Code using animationFrameScheduler breaks in Firefox when using Sentry.',
          },
          {
            name: 'react-router-dom',
            importNames: ['Link'],
            message: 'Use the <Link /> component from @sourcegraph/wildcard instead.',
          },
          {
            name: '@sourcegraph/wildcard',
            importNames: ['Tooltip'],
            message:
              'Please ensure there is only a single `<Tooltip />` component present in the React tree. To display a specific tooltip, you can add the `data-tooltip` attribute to the relevant element.',
          },
          {
            name: 'zustand',
            importNames: ['default'],
            message:
              'Our Zustand stores should be created in a single place. Create this store in client/web/src/stores',
          },
        ],
        patterns: [
          {
            group: ['**/enterprise/*'],
            message: `The OSS product may not pull in any code from the enterprise codebase, to stay a 100% open-source program.

See https://handbook.sourcegraph.com/community/faq#is-all-of-sourcegraph-open-source for more information.`,
          },
        ],
      },
    ],
    'react/forbid-elements': [
      'error',
      {
        forbid: [
          {
            element: 'form',
            message:
              'Use the Form component in src/components/Form.tsx instead of the native HTML form element to get proper form validation feedback',
          },
          {
            element: 'select',
            message: 'Use the <Select /> component from @sourcegraph/wildcard instead.',
          },
          {
            element: 'textarea',
            message: 'Use the <TextArea /> component from @sourcegraph/wildcard instead.',
          },
          {
            element: 'a',
            message: 'Use the <Link /> component from @sourcegraph/wildcard instead.',
          },
        ],
      },
    ],
    '@sourcegraph/wildcard/forbid-class-name': [
      'error',
      {
        forbid: [
          {
            className: 'badge',
            message: 'Use the <Badge /> component from @sourcegraph/wildcard instead.',
          },
        ],
      },
    ],
    'react/jsx-no-target-blank': ['error', { allowReferrer: true }],
  },
  overrides: [
    {
      files: ['*.d.ts'],
      rules: {
        'no-restricted-imports': 'off',
      },
    },
    {
      files: '*.story.tsx',
      rules: {
        'react/forbid-dom-props': 'off',
        'import/no-default-export': 'off',
      },
    },
  ],
}

module.exports = config
