module.exports = {
  ci: {
    collect: {
      url: ['http://localhost:3443/search'],
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
