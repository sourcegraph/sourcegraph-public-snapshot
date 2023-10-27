import { describe, expect, test } from '@jest/globals'

import { createRectangle, getIntersection } from './rectangle'

describe('rectangle should calculate intersection', () => {
    const intersectionCases = [
        [createRectangle(500, 500, 200, 200), createRectangle(0, 0, 600, 600), createRectangle(500, 500, 100, 100)],
        [createRectangle(0, 0, 200, 200), createRectangle(10, 10, 600, 600), createRectangle(10, 10, 190, 190)],
        [createRectangle(0, 0, 200, 200), createRectangle(210, 210, 600, 600), createRectangle(0, 0, 0, 0)],
    ]

    test.each(intersectionCases)('with %s and %s rectangles', (a, b, expected) => {
        expect(getIntersection(a, b)).toStrictEqual(expected)
    })
})
