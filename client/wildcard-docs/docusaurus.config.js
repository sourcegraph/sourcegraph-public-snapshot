const darkCodeTheme = require('prism-react-renderer/themes/dracula');
const lightCodeTheme = require('prism-react-renderer/themes/github');

// With JSDoc @type annotations, IDEs can provide config autocompletion
/** @type {import('@docusaurus/types').DocusaurusConfig} */
(module.exports = {
  title: 'Wildcard Library Design System',
  tagline: 'A collection of design-approved reusable components that are suitable for use within the Sourcegraph codebase.',
  url: 'https://your-docusaurus-test-site.com',
  baseUrl: '/',
  onBrokenLinks: 'throw',
  onBrokenMarkdownLinks: 'warn',
  favicon: 'img/logo.svg',
  organizationName: 'Sourcegraph', // Usually your GitHub org/user name.
  projectName: 'wildcard-docs', // Usually your repo name.
  // plugins: [
  //   [
  //     'docusaurus-plugin-typedoc',

  //     // Plugin / TypeDoc options
  //     {
  //       entryPoints: '../wildcard/src/index.ts',
  //       tsconfig: '../wildcard/tsconfig.json',
  //       entryPointStrategy: 'Expand',
  //       exclude: '../wildcard/+(.test|.story|.module.scss).tsx',
  //       allReflectionsHaveOwnDocument: true
  //     },
  //   ],
  // ],
  presets: [
    [
      '@docusaurus/preset-classic',
      /** @type {import('@docusaurus/preset-classic').Options} */
      ({
        docs: {
          sidebarPath: require.resolve('./sidebars.js'),
          editUrl: 'https://github.com/sourcegraph/sourcegraph/wildcard-docs/edit/main/website/',
          path: './docs',
          routeBasePath: '/'
        },
        blog: false,
        theme: {
          customCss: require.resolve('./src/css/custom.css'),
        },
      }),
    ],
  ],

  themeConfig:
    /** @type {import('@docusaurus/preset-classic').ThemeConfig} */
    ({
      navbar: {
        title: 'Wildcard Library',
        logo: {
          alt: 'Wildcard Library',
          src: 'img/logo.svg',
        },
      },
      prism: {
        theme: lightCodeTheme,
        darkTheme: darkCodeTheme,
      },
    }),
});
