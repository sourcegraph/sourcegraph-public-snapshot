import { mdiFileCodeOutline, mdiFilePngBox, mdiLanguageJavascript } from '@mdi/js'
import { describe, expect, it } from 'vitest'

import { ALL_LANGUAGES } from '@sourcegraph/common'

import { getFileIconInfo, FILE_ICONS_BY_LANGUAGE } from './language-icons'

describe('checkValidLanguageNames', () => {
    const allLanguagesSet = new Set(ALL_LANGUAGES)
    for (const [languageName, _] of FILE_ICONS_BY_LANGUAGE) {
        it(languageName, () => {
            expect(allLanguagesSet.has(languageName)).toBeTruthy()
        })
    }
})

describe('getFileIconInfo', () => {
    const tests: {
        name: string
        file: string
        language: string
        expectedSvgPath: string | undefined
        expectedIsTest: boolean
    }[] = [
        {
            name: 'check that png works',
            file: 'myfile.png',
            language: '',
            expectedSvgPath: mdiFilePngBox,
            expectedIsTest: false,
        },
        {
            name: 'works with simple file name',
            file: 'my-file.js',
            language: 'JavaScript',
            expectedSvgPath: mdiLanguageJavascript,
            expectedIsTest: false,
        },
        {
            name: 'check fallback behavior',
            file: 'placeholder',
            language: 'Vim Script',
            expectedSvgPath: mdiFileCodeOutline,
            expectedIsTest: false,
        },
        {
            name: 'check unknown language',
            file: 'my-file.test.unknown',
            language: 'Unknown',
            expectedSvgPath: undefined,
            expectedIsTest: true,
        },
    ]

    for (const t of tests) {
        it(t.name, () => {
            const iconInfo = getFileIconInfo(t.file, t.language)
            expect(iconInfo?.svg.path).toBe(t.expectedSvgPath)
        })
    }
})
