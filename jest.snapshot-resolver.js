// @ts-check

const path = require('path')

// TODO(bazel): bazel runs tests on the pre-compiled .js files, non-bazel runs on .ts files.
// This snapshot resolver edits the pre-compiled .js to snapshots assuming a .tsx extension.
// This can be removed and snapshot files renamed to .js when non-bazel is removed.
// NOTE: this assumes all snapshot tests are in .tsx files, not .ts or .jsx and will not work for non-.tsx files.

const SNAPSHOT_EXTENSION = '.snap'
const TEST_EXTENSION = '.tsx'

module.exports = {
  resolveSnapshotPath: testPath =>
    path.join(
      path.join(path.dirname(testPath), '__snapshots__'),
      path.basename(testPath).replace('.js', TEST_EXTENSION) + SNAPSHOT_EXTENSION
    ),

  resolveTestPath: snapshotPath =>
    path.join(
      path.dirname(path.dirname(snapshotPath)),
      path.basename(snapshotPath, SNAPSHOT_EXTENSION).replace('.js', TEST_EXTENSION)
    ),

  testPathForConsistencyCheck: path.posix.join('consistency_check', '__tests__', 'example.test.tsx'),
}
