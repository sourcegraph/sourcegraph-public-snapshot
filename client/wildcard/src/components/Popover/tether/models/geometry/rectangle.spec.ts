import { createRectangle, intersection } from './rectangle'

describe('rectangle should calculate intersection', () => {
    const intersectionCases = [
        [createRectangle(500, 500, 200, 200), createRectangle(0, 0, 600, 600), createRectangle(500, 500, 100, 100)],
        [createRectangle(0, 0, 200, 200), createRectangle(10, 10, 600, 600), createRectangle(10, 10, 190, 190)],
    ]

    test.each(intersectionCases)('with %s and %s rectangles', (a, b, expected) => {
        expect(intersection(a, b)).toStrictEqual(expected)
    })
})
