module.exports = {
  ci: {
    collect: {
      url: ['http://localhost:3443/'],
      startServerCommand: 'yarn workspace @sourcegraph/web serve:prod',
    },
    upload: {
      target: 'temporary-public-storage',
    },
  },
}
