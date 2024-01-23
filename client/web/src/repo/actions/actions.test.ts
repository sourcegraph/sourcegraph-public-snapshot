import { expect, describe, test } from 'vitest'

import { createUrl } from './CopyPermalinkAction'

describe('createURL', () => {
    test('should return the correct URL given the rooturl and path', () => {
        const url = 'http://localhost:3080'
        const path = '/api/v1?q=foo#bar'
        const wantUrl = createUrl(url, path)
        const gotUrl = 'http://localhost:3080/api/v1?q=foo#bar'

        expect(wantUrl).toBe(gotUrl)
    })
})
