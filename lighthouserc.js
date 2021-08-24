module.exports = {
  ci: {
    collect: {
      url: [
        'http://localhost:3443/search',
        'http://localhost:3443/search?q=repo:sourcegraph/sourcegraph&patternType=literal',
        'http://localhost:3443/github.com/sourcegraph/sourcegraph',
      ],
      startServerCommand: 'yarn workspace @sourcegraph/web serve:prod',
      settings: {
        preset: 'desktop',
      },
    },
    upload: {
      target: 'temporary-public-storage',
    },
    // assert: {
    //   preset: 'lighthouse:no-pwa',
    // },
  },
}
