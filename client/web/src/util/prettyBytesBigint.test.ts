import { describe, expect, test } from '@jest/globals'

import { prettyBytesBigint } from './prettyBytesBigint'

describe('prettyBytesBigint', () => {
    test('pretty prints 1.23 Gb', () => {
        expect(prettyBytesBigint(BigInt(1231234560))).toBe('1.23 GB')
    })
    test('pretty prints 1.02 Gb', () => {
        expect(prettyBytesBigint(BigInt(1021234567))).toBe('1.02 GB')
    })
    test('pretty prints 500.50 Gb', () => {
        expect(prettyBytesBigint(BigInt(500500001337))).toBe('500.50 GB')
    })
    test('pretty prints 50.05 Tb', () => {
        expect(prettyBytesBigint(BigInt(50055550000000))).toBe('50.05 TB')
    })
    test('pretty prints 50 Gb', () => {
        expect(prettyBytesBigint(BigInt(50005555000))).toBe('50 GB')
    })
    test('pretty prints 1 Mb', () => {
        expect(prettyBytesBigint(BigInt(1230000))).toBe('1 MB')
    })
    test('pretty prints 999 Mb', () => {
        expect(prettyBytesBigint(BigInt(999123000))).toBe('999 MB')
    })
})
