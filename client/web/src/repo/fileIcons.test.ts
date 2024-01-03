import { mdiFilePngBox, mdiLanguageJavascript } from '@mdi/js'
import { describe, expect, it } from 'vitest'

import { ALL_LANGUAGES } from '@sourcegraph/shared/src/search/query/languageFilter'

import { isProbablyTestFile, getFileIconInfo, onlyForTesting } from './fileIcons'

describe('checkValidLanguageNames', () => {
    const allLanguagesSet = new Set(ALL_LANGUAGES)
    for (const [languageName, _] of onlyForTesting.FILE_ICONS_BY_LANGUAGE) {
        it(languageName, () => {
            expect(allLanguagesSet.has(languageName)).toBeTruthy()
        })
    }
})

describe('getFileIconInfo', () => {
    const tests: {
        name: string
        file: string
        languages: string[]
        expectedSvgPath: string | undefined
        expectedIsTest: boolean
    }[] = [
        {
            name: 'check that png works',
            file: 'myfile.png',
            languages: [],
            expectedSvgPath: mdiFilePngBox,
            expectedIsTest: false,
        },
        {
            name: 'works with simple file name',
            file: 'my-file.js',
            languages: ['JavaScript'],
            expectedSvgPath: mdiLanguageJavascript,
            expectedIsTest: false,
        },
        {
            name: 'check fallback behavior',
            file: 'placeholder',
            languages: ['Vim Script'],
            expectedSvgPath: onlyForTesting.DEFAULT_CODE_FILE_ICON.path,
            expectedIsTest: false,
        },
        {
            name: 'check unknown language',
            file: 'my-file.test.unknown',
            languages: ['Unknown'],
            expectedSvgPath: undefined,
            expectedIsTest: true,
        },
    ]

    for (const t of tests) {
        it(t.name, () => {
            const iconInfo = getFileIconInfo(t.file, t.languages)
            expect(iconInfo?.svg.path).toBe(t.expectedSvgPath)

            const isLikelyTest = isProbablyTestFile(t.file)
            expect(isLikelyTest).toBe(t.expectedIsTest)
        })
    }
})

describe('containsTest', () => {
    const tests: {
        name: string
        file: string
        expected: boolean
    }[] = [
        {
            name: 'returns true if "test_" exists in file name',
            file: 'test_myfile.go',
            expected: true,
        },
        {
            name: 'returns true if "_test" exists in file name',
            file: 'myfile_test.go',
            expected: true,
        },
        {
            name: 'returns true if "_spec" exists in file name',
            file: 'myfile_spec.go',
            expected: true,
        },
        {
            name: 'returns true if "spec_" exists in file name',
            file: 'spec_myfile.go',
            expected: true,
        },
        {
            name: 'works with sub-extensions',
            file: 'myreactcomponent.test.tsx',
            expected: true,
        },
        {
            name: 'returns false if not a test file',
            file: 'mytestcomponent.java',
            expected: false,
        },
    ]

    for (const t of tests) {
        it(t.name, () => {
            expect(isProbablyTestFile(t.file)).toBe(t.expected)
        })
    }
})
