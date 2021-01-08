import { decorationAttachmentStyleForTheme, decorationStyleForTheme, fileDecorationColorForTheme } from './decoration'

describe('decorationStyleForTheme', () => {
    const FIXTURE_RANGE = { start: { line: 1, character: 2 }, end: { line: 3, character: 4 } }

    test('supports no theme overrides', () =>
        expect(decorationStyleForTheme({ range: FIXTURE_RANGE, backgroundColor: 'red' }, true)).toEqual({
            backgroundColor: 'red',
        }))

    test('applies light theme overrides', () =>
        expect(
            decorationStyleForTheme(
                { range: FIXTURE_RANGE, backgroundColor: 'red', light: { backgroundColor: 'blue' } },
                true
            )
        ).toEqual({
            backgroundColor: 'blue',
        }))

    test('applies dark theme overrides', () =>
        expect(
            decorationStyleForTheme(
                {
                    range: FIXTURE_RANGE,
                    backgroundColor: 'red',
                    light: { backgroundColor: 'blue' },
                    dark: { backgroundColor: 'green' },
                },
                false
            )
        ).toEqual({
            backgroundColor: 'green',
        }))
})

describe('decorationAttachmentStyleForTheme', () => {
    test('supports no theme overrides', () =>
        expect(decorationAttachmentStyleForTheme({ color: 'red' }, true)).toEqual({ color: 'red' }))

    test('applies light theme overrides', () =>
        expect(decorationAttachmentStyleForTheme({ color: 'red', light: { color: 'blue' } }, true)).toEqual({
            color: 'blue',
        }))

    test('applies dark theme overrides', () =>
        expect(
            decorationAttachmentStyleForTheme(
                { color: 'red', light: { color: 'blue' }, dark: { color: 'green' } },
                false
            )
        ).toEqual({
            color: 'green',
        }))
})

describe('fileDecorationColorForTheme', () => {
    test('supports no theme overrides', () => {
        expect(
            fileDecorationColorForTheme(
                {
                    contentText: '',
                    color: 'red',
                },
                false
            )
        ).toEqual('red')
    })

    test('applies light theme overrides', () => {
        expect(
            fileDecorationColorForTheme(
                {
                    contentText: '',
                    color: 'red',
                    light: {
                        color: 'blue',
                    },
                },
                true
            )
        ).toEqual('blue')
    })

    test('applies dark theme overrides', () => {
        expect(
            fileDecorationColorForTheme(
                {
                    contentText: '',
                    color: 'red',
                    dark: {
                        color: 'green',
                    },
                },
                false
            )
        ).toEqual('green')
    })

    test('applies selected color overrides', () => {
        expect(
            fileDecorationColorForTheme(
                {
                    contentText: '',
                    color: 'red',
                    activeColor: 'orange',
                },
                false,
                true
            )
        ).toEqual('orange')
    })

    test('applies selected color for themes', () => {
        expect(
            fileDecorationColorForTheme(
                {
                    contentText: '',
                    color: 'red',
                    activeColor: 'orange',
                    dark: {
                        activeColor: 'teal',
                    },
                },
                false,
                true
            )
        ).toEqual('teal')
    })
})
