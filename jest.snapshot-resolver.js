// @ts-check

const path = require('path')

// TODO(bazel): drop when non-bazel removed.
const IS_BAZEL = !!(process.env.JS_BINARY__TARGET || process.env.BAZEL_BINDIR || process.env.BAZEL_TEST)

// TODO(bazel): bazel runs tests on the pre-compiled .js files, non-bazel runs on .tsx files.
// This snapshot resolver edits the pre-compiled .js to snapshots assuming a .tsx extension.
// This can be removed and snapshot files renamed to .js when non-bazel is removed.
// NOTE: this assumes all snapshot tests are in .tsx files, not .ts or .jsx and will not work for non-.tsx files.

const SNAPSHOT_EXTENSION = '.tsx.snap'
const TEST_EXTENSION = IS_BAZEL ? '.js' : '.tsx'

module.exports = {
  resolveSnapshotPath: testPath =>
    path.join(
      path.join(path.dirname(testPath), '__snapshots__'),
      path.basename(testPath).replace(TEST_EXTENSION, SNAPSHOT_EXTENSION)
    ),

  resolveTestPath: snapshotPath =>
    path.join(
      path.dirname(path.dirname(snapshotPath)),
      path.basename(snapshotPath).replace(SNAPSHOT_EXTENSION, TEST_EXTENSION)
    ),
  testPathForConsistencyCheck: path.posix.join(__dirname, 'consistency_check', '__tests__', 'example.test.tsx'),
}
