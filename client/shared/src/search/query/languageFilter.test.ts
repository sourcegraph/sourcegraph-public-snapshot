import { languageCompletion, popularLanguages, allLanguages } from './languageFilter'
import { Literal, createLiteral } from './token'

const create = (value: string): Literal => createLiteral(value, { start: 0, end: 0 })

describe('languageCompletion', () => {
    test('suggest popular languages', () => {
        expect(languageCompletion(undefined)).toStrictEqual(popularLanguages)
        expect(languageCompletion(create(''))).toStrictEqual(popularLanguages)
    })

    test('suggest all languages', () => {
        expect(languageCompletion(create('c'))).toStrictEqual(allLanguages)
    })
})
