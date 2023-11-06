import { describe, expect, it } from '@jest/globals'

import { splitExtensionID } from './extension'

describe('splitExtensionID', () => {
    it('splits extensionID with host', () => {
        expect(splitExtensionID('sourcegraph.example.com/bob/myextension')).toStrictEqual({
            host: 'sourcegraph.example.com',
            publisher: 'bob',
            name: 'myextension',
        })
    })
    it('splits extensionID without host', () => {
        expect(splitExtensionID('alice/myextension')).toStrictEqual({
            publisher: 'alice',
            name: 'myextension',
            isSourcegraphExtension: false,
        })
    })
})
