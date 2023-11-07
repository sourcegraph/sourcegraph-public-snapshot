// @ts-check

// A minimal babel config only for jest transformations.
// All typescript and react transformations are done by previous
// bazel build rules, so we only need to do jest transformations here.

// TODO(bazel): drop when non-bazel removed.
if (!(process.env.JS_BINARY__TARGET || process.env.BAZEL_BINDIR || process.env.BAZEL_TEST)) {
  throw new Error(__filename + ' is only for use with Bazel')
}

/** @type {import('@babel/core').ConfigFunction} */
module.exports = api => {
  api.cache.forever()

  return {
    presets: [
      [
        '@babel/preset-env',
        {
          targets: {
            // We only run jest tests in node. All the browser related transformations
            // are already completed on the previous transpilation step.
            node: '16',
          },
        },
      ],
    ],
  }
}
