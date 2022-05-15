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
    project: __dirname + '/tsconfig.eslint.json',
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
            name: 'reactstrap',
            importNames: ['Form'],
            message: 'Use the <Form /> component from @sourcegraph/branded package instead',
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
          {
            name: 'reactstrap',
            message:
              'Please use components from the Wildcard component library instead. We work on removing `reactstrap` dependency.',
          },
        ],
        patterns: [
          {
            group: ['**/enterprise/*'],
            message: `The OSS product may not pull in any code from the enterprise codebase, to stay a 100% open-source program.

See https://handbook.sourcegraph.com/community/faq#is-all-of-sourcegraph-open-source for more information.`,
          },
          {
            group: [
              '@sourcegraph/*/src/*',
              '!@sourcegraph/branded/src/*',
              '!@sourcegraph/shared/src/*',
              '!@sourcegraph/web/src/SourcegraphWebApp.scss',
            ],
            message:
              'Imports from package internals are banned. Add relevant export to the entry point of the package to import it from the outside world.',
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
          {
            element: 'h1',
            message: 'Use the <Typography.H1 /> component from @sourcegraph/wildcard instead.',
          },
          {
            element: 'h2',
            message: 'Use the <Typography.H2 /> component from @sourcegraph/wildcard instead.',
          },
          {
            element: 'h3',
            message: 'Use the <Typography.H3 /> component from @sourcegraph/wildcard instead.',
          },
          {
            element: 'h4',
            message: 'Use the <Typography.H4 /> component from @sourcegraph/wildcard instead.',
          },
          {
            element: 'h5',
            message: 'Use the <Typography.H5 /> component from @sourcegraph/wildcard instead.',
          },
          {
            element: 'h6',
            message: 'Use the <Typography.H6 /> component from @sourcegraph/wildcard instead.',
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
          {
            className: 'icon-inline',
            message: 'Use the <Icon /> component from @sourcegraph/wildcard instead.',
          },
        ],
      },
    ],
    'react/jsx-no-target-blank': ['error', { allowReferrer: true }],
    'no-restricted-syntax': [
      'warn',
      {
        selector: 'CallExpression[callee.name="useLocalStorage"]',
        message:
          'Consider using useTemporarySetting instead of useLocalStorage so settings are synced when users log in elsewhere. More info at https://docs.sourcegraph.com/dev/background-information/web/temporary_settings',
      },
    ],
    // https://reactjs.org/blog/2020/09/22/introducing-the-new-jsx-transform.html#eslint
    'react/jsx-uses-react': 'off',
    'react/react-in-jsx-scope': 'off',
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
