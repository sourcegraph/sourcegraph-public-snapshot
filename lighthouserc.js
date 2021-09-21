// @ts-check

const config = {
  ci: {
    collect: {
      url: [
        'http://localhost:3443/search', // Homepage
        'http://localhost:3443/search?q=context:global+repo:sourcegraph/sourcegraph+file:lighthouserc.js&patternType=literal', // Search result
        'http://localhost:3443/github.com/sourcegraph/sourcegraph/-/blob/package.json', // File blob
      ],
      startServerCommand: 'yarn workspace @sourcegraph/web serve:prod',
      settings: {
        preset: 'desktop',
        // These audits are not currently supported by the local production server.
        // TODO: Check why errors in console is here
        skipAudits: ['meta-description', 'is-on-https', 'uses-http2'],
        chromeFlags: '--no-sandbox',
      },
    },
    upload: {
      target: 'temporary-public-storage',
    },
  },
}

module.exports = config
