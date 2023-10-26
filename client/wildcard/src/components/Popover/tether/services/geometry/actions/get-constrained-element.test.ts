import { describe, expect, test } from '@jest/globals'

import { createRectangle } from '../../../models/geometry/rectangle'

import { getConstrainedElement } from './get-constrained-element'

describe('getConstrainedElement', () => {
    const testCases = [
        [createRectangle(500, 500, 200, 200), createRectangle(0, 0, 600, 600), createRectangle(400, 400, 200, 200)],
        [createRectangle(0, 0, 200, 200), createRectangle(10, 10, 600, 600), createRectangle(10, 10, 200, 200)],
    ]

    test.each(testCases)('should return correct result with overflowed content', (a, b, expected) => {
        expect(getConstrainedElement(a, b)).toStrictEqual(expected)
    })
})
