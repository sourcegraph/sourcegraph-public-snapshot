import { applyExtensionsQuery } from './ExtensionRegistry'

describe('applyExtensionsQuery', () => {
    test('#installed', () =>
        expect(
            applyExtensionsQuery('#installed', { extensions: { a: true, b: false } }, [
                { extensionID: 'a' },
                { extensionID: 'b' },
                { extensionID: 'c' },
            ]).map(({ extensionID }) => extensionID)
        ).toEqual(['a', 'b']))

    test('#enabled', () =>
        expect(
            applyExtensionsQuery('#enabled', { extensions: { a: true, b: false } }, [
                { extensionID: 'a' },
                { extensionID: 'b' },
                { extensionID: 'c' },
            ]).map(({ extensionID }) => extensionID)
        ).toEqual(['a']))

    test('#disabled', () =>
        expect(
            applyExtensionsQuery('#disabled', { extensions: { a: true, b: false } }, [
                { extensionID: 'a' },
                { extensionID: 'b' },
                { extensionID: 'c' },
            ]).map(({ extensionID }) => extensionID)
        ).toEqual(['b']))
})
