import { expect, describe, test } from 'vitest'

import { createUrl } from './CopyPermalinkAction'

describe('URL', () => {
    test('should return the correct URL', () => {
        let rooturl = 'http://localhost:3000'
        let path = '/api/v1'
        let combined = createUrl(rooturl, path)

        expect(combined).toBe('http://localhost:3000/api/v1')
    })
})
