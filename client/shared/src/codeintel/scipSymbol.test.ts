import { describe, expect, test } from 'vitest'

import { type SCIPSymbol, formatSymbol, parseSymbol } from './scipSymbol'

describe('parseSymbolName', () => {
    test('roundtrip', () => {
        const testCases: [string, string][] = [
            ['scip-java . . . Dude#lol!waow.', 'scip-java . . . Dude#lol!waow.'],
            ['scip  java . . . Dude#lol!waow.', 'scip  java . . . Dude#lol!waow.'],
            ['scip  java . . . `Dude```#`lol`!waow.', 'scip  java . . . `Dude```#lol!waow.'],
            ['local 1', 'local 1'],
            [
                'rust-analyzer cargo test_rust_dependency 0.1.0 println!',
                'rust-analyzer cargo test_rust_dependency 0.1.0 println!',
            ],
        ]

        for (const testCase of testCases) {
            const parsed = parseSymbol(testCase[0])
            const formatted = formatSymbol(parsed)
            expect(formatted).toEqual(testCase[1])
        }
    })

    test('parse', () => {
        interface test {
            symbol: string
            expected: SCIPSymbol
        }
        const tests: test[] = [
            { symbol: 'local a', expected: { kind: 'local', localID: 'a' } },
            {
                symbol: 'a b c d method().',
                expected: {
                    kind: 'nonlocal',
                    scheme: 'a',
                    package: {
                        manager: 'b',
                        name: 'c',
                        version: 'd',
                    },
                    descriptors: [{ name: 'method', kind: 'method' }],
                },
            },
            // Backtick-escaped descriptor
            {
                symbol: 'a b c d `e f`.',
                expected: {
                    kind: 'nonlocal',
                    scheme: 'a',
                    package: {
                        manager: 'b',
                        name: 'c',
                        version: 'd',
                    },
                    descriptors: [{ name: 'e f', kind: 'term' }],
                },
            },
            // Space-escaped package name
            {
                symbol: 'a b  c d e f.',
                expected: {
                    kind: 'nonlocal',
                    scheme: 'a',
                    package: {
                        manager: 'b c',
                        name: 'd',
                        version: 'e',
                    },
                    descriptors: [{ name: 'f', kind: 'term' }],
                },
            },
            {
                symbol: 'lsif-java maven package 1.0.0 java/io/File#Entry.method(+1).(param)[TypeParam]',
                expected: {
                    kind: 'nonlocal',
                    scheme: 'lsif-java',
                    package: { manager: 'maven', name: 'package', version: '1.0.0' },
                    descriptors: [
                        { name: 'java', kind: 'namespace' },
                        { name: 'io', kind: 'namespace' },
                        { name: 'File', kind: 'type' },
                        { name: 'Entry', kind: 'term' },
                        { name: 'method', disambiguator: '+1', kind: 'method' },
                        { name: 'param', kind: 'parameter' },
                        { name: 'TypeParam', kind: 'typeParameter' },
                    ],
                },
            },
            {
                symbol: 'rust-analyzer cargo std 1.0.0 macros/println!',
                expected: {
                    kind: 'nonlocal',
                    scheme: 'rust-analyzer',
                    package: { manager: 'cargo', name: 'std', version: '1.0.0' },
                    descriptors: [
                        { name: 'macros', kind: 'namespace' },
                        { name: 'println', kind: 'macro' },
                    ],
                },
            },
        ]

        for (const testCase of tests) {
            expect(parseSymbol(testCase.symbol)).toEqual(testCase.expected)
        }
    })
})
