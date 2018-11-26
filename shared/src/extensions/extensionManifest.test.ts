import assert from 'assert'
import { isErrorLike } from '../util/errors'
import { parseExtensionManifestOrError } from './extensionManifest'

describe('parseExtensionManifestOrError', () => {
    it('parses valid input', () =>
        assert.deepStrictEqual(parseExtensionManifestOrError('{"url":"a","activationEvents":["*"]}'), {
            url: 'a',
            activationEvents: ['*'],
        }))
    it('returns an error value for invalid JSONC', () => {
        const value = parseExtensionManifestOrError('.')
        assert.ok(isErrorLike(value))
    })
    it('returns an error value for valid JSONC but invalid data', () => {
        const value = parseExtensionManifestOrError('{"url":"a"}')
        assert.ok(isErrorLike(value))
    })
})
