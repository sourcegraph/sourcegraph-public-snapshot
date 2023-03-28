import type { Manifest } from 'webextension-polyfill';
import pkg from '../package.json';

const manifest: Manifest.WebExtensionManifest = {
  manifest_version: 3,
  name: pkg.displayName,
  version: pkg.version,
  description: pkg.description,
  permissions: ['tabs', 'contextMenus', 'storage', 'scripting', 'activeTab'],
  host_permissions: [
    'https://*/*',
    'https://sourcegraph.com/*',
    'https://github.com/*',
    'https://gitlab.com/*',
    'https://*.sgdev.org/*',
    'https://sourcegraph.sourcegraph.com/*',
    'https://stackoverflow.com/questions/*',
  ],
  options_ui: {
    page: 'src/pages/options/index.html',
  },
  background: {
    service_worker: 'src/pages/background/index.js',
    type: 'module',
  },
  action: {
    default_popup: 'src/pages/popup/index.html',
    default_icon: 'cody.png',
  },
  icons: {
    '128': 'cody.png',
  },
  content_scripts: [
    {
      matches: [
        'https://*/*',
        'https://sourcegraph.com/*',
        'https://github.com/*',
        'https://gitlab.com/*',
        'https://*.sgdev.org/*',
        'https://sourcegraph.sourcegraph.com/*',
        'https://stackoverflow.com/questions/*',
      ],
      js: ['src/pages/content/index.js'],
      css: ['contentStyle.css'],
      run_at: 'document_end',
    },
  ],
  devtools_page: 'src/pages/devtools/index.html',
  web_accessible_resources: [
    {
      resources: ['contentStyle.css', 'cody.png', 'overlay/index.html'],
      matches: [],
    },
  ],
};

export default manifest;
