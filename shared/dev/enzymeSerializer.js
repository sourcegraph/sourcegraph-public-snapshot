import { createSerializer } from 'enzyme-to-json';
import { replaceHistoryObject } from '../src/util/enzymeSnapshotModifiers'

module.exports = createSerializer({ map: replaceHistoryObject })

