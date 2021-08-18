const darkCodeTheme = require('prism-react-renderer/themes/dracula');
const lightCodeTheme = require('prism-react-renderer/themes/github');

/** @type {import('@docusaurus/types').DocusaurusConfig} */
module.exports = {
  title: 'Wildcard Components Library',
  tagline: 'Design system by Sourcegraph',
  url: 'https://your-docusaurus-test-site.com',
  baseUrl: '/',
  onBrokenLinks: 'throw',
  onBrokenMarkdownLinks: 'warn',
  favicon: 'img/favicon.ico',
  organizationName: 'sourcegraph', // Usually your GitHub org/user name.
  projectName: 'wcl', // Usually your repo name.
  themeConfig: {
    navbar: {
      title: 'Wildcard Components Library',
      logo: {
        alt: 'Wildcard Components Library - Sourcegraph',
        src: 'img/logo.svg',
      },
      items: [
        {
          type: 'doc',
          docId: 'intro',
          position: 'left',
          label: 'Components',
        },
        { to: '/components', label: 'Blog', position: 'left' },
        {
          href: 'https://github.com/sourcegraph/sourcegraph',
          label: 'GitHub',
          position: 'right',
        },
      ],
    },
    prism: {
      theme: lightCodeTheme,
      darkTheme: darkCodeTheme,
    },
  },
  presets: [
    [
      '@docusaurus/preset-classic',
      {
        docs: {
          sidebarPath: require.resolve('./sidebars.ts'),
          // Please change this to your repo.
          editUrl:
            'https://github.com/sourcegraph/sourcegraph/edit/master/website/blog/',
        },
        theme: {
          customCss: require.resolve('./src/css/custom.css'),
        },
      },
    ],
  ],
};
