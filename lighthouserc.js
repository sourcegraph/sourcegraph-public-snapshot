// @ts-check

const config = {
  ci: {
    collect: {
      // Note: We override this URL in CI through ./dev/ci/yarn-lighthouse.sh
      url: 'https://sourcegraph.test:3443/',
      startServerCommand: 'yarn workspace @sourcegraph/web serve:prod',
      settings: {
        preset: 'desktop',
        chromeFlags: '--no-sandbox',
        // We skip a series of audits that are not currently supported by the local server
        skipAudits: [
          // SEO: Normally enabled dynamically for different paths in the production server
          'meta-description',
          // SEO: Robots.txt file isn't served locally
          'robots-txt',
        ],
      },
    },
    upload: {
      target: 'temporary-public-storage',
    },
  },
}

module.exports = config
