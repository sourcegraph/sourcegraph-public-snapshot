// @ts-check

const config = {
  ci: {
    collect: {
      url: ['http://localhost:3443/search'],
      startServerCommand: 'yarn workspace @sourcegraph/web serve:prod',
      psiStrategy: 'desktop',
      settings: {
        preset: 'desktop',
      },
    },
    upload: {
      target: 'temporary-public-storage',
    },
    assert: {
      assertions: {
        'categories:pwa': 'off',
        // Not supported by local production server yet
        'meta-description': 'off',
        // Not supported by local production server yet
        'is-on-https': 'off',
        'uses-http2': 'off',
        // TODO: Check why - ideally re-enable
        'errors-in-console': 'off',
      },
    },
  },
}

module.exports = config
