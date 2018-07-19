import * as assert from 'assert'
import { getPathExtension } from '.'

describe('util/index', () => {
    describe('getPathExtension', () => {
        it('returns extension if normal path', () => {
            assert.strictEqual(getPathExtension('/foo/baz/bar.go'), 'go')
        })

        it('returns empty string if no extension', () => {
            assert.strictEqual(getPathExtension('README'), '')
        })

        it('returns empty string if hidden file with no extension', () => {
            assert.strictEqual(getPathExtension('.gitignore'), '')
        })

        it('returns extension for path with multiple dot separators', () => {
            assert.strictEqual(getPathExtension('.baz.bar.go'), 'go')
        })
    })
})
