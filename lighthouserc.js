// @ts-check

const config = {
  ci: {
    collect: {
      url: ['http://localhost:3443/search'],
      startServerCommand: 'yarn workspace @sourcegraph/web serve:prod',
      settings: {
        preset: 'desktop',
        // These audits are not currently supported by the local production server.
        // TODO: Check why errors in console is here
        skipAudits: ['meta-description', 'is-on-https', 'uses-http2', 'errors-in-console'],
      },
    },
    upload: {
      target: 'temporary-public-storage',
    },
  },
}

module.exports = config
