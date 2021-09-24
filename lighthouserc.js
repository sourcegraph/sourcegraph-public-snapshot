// @ts-check

const config = {
  ci: {
    collect: {
      // Note: We ovveride this URL in CI through ./dev/ci/yarn-lighthouse.sh
      url: 'http://localhost:3443/',
      startServerCommand: 'yarn workspace @sourcegraph/web serve:prod',
      settings: {
        preset: 'desktop',
        chromeFlags: '--no-sandbox',
        // Skip audits on features that are not currently supported by the local production server.
        skipAudits: ['meta-description', 'is-on-https', 'uses-http2', 'errors-in-console'],
      },
    },
    upload: {
      target: 'temporary-public-storage',
    },
  },
}

module.exports = config
