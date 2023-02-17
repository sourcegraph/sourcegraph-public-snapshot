// @ts-check

const path = require('path')

const SNAPSHOT_EXTENSION = '.snap'
const TEST_EXTENSION = '.tsx'

module.exports = {
  resolveSnapshotPath: testPath => {
    const filename = path.parse(testPath).name + SNAPSHOT_EXTENSION

    return path.join(path.join(path.dirname(testPath), '__snapshots__'), filename)
  },

  resolveTestPath: snapshotPath => {
    const filename = path.parse(snapshotPath).name + TEST_EXTENSION

    return path.join(path.dirname(path.dirname(snapshotPath)), filename)
  },

  testPathForConsistencyCheck: path.posix.join('consistency_check', '__tests__', 'example.test.tsx'),
}
