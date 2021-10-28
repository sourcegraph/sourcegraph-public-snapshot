// @ts-check

/** @type {import('svgo').OptimizeOptions} */
const config = {
  plugins: [
    {
      name: 'preset-default',
      params: {
        overrides: {
          // Do not remove viewBox, we need it to accurately scale SVGs: https://github.com/svg/svgo/issues/1128
          removeViewBox: false,
        },
      },
    },
  ],
}

module.exports = config
