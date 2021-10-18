// @ts-check

const config = {
  ci: {
    collect: {
      url: [
        'http://localhost:3080',
        'http://localhost:3080/search?q=repo:sourcegraph/lighthouse-ci-test-repository+file:index.js',
        'http://localhost:3080/github.com/sourcegraph/lighthouse-ci-test-repository',
        'http://localhost:3080/github.com/sourcegraph/lighthouse-ci-test-repository/-/blob/index.js',
      ],
      startServerCommand: 'yarn workspace @sourcegraph/web serve:prod',
      settings: {
        preset: 'desktop',
        chromeFlags: '--no-sandbox',
        // We skip a series of audits that are not currently supported by the local server
        skipAudits: [
          // SEO: Normally enabled dynamically for different paths in the production server
          'meta-description',
          // Best practices: HTTPS currently disabled in local server: https://github.com/sourcegraph/sourcegraph/issues/21869
          'is-on-https',
          'uses-http2',
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
