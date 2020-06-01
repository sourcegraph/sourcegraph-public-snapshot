import { createSerializer } from 'enzyme-to-json';
import { replaceHistoryObject } from '../src/util/enzymeSnapshotModifiers'

function config() {
  return createSerializer({ map: replaceHistoryObject })
}

module.exports = config()

