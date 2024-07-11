import { PiFilePngLight } from 'react-icons/pi'
import { SiJavascript } from 'react-icons/si'
import { describe, expect, it } from 'vitest'

import { getFileIconInfo } from './language-icons'

describe('getFileIconInfo', () => {
    const tests = [
        {
            name: 'check that png works',
            file: 'myfile.png',
            language: '',
            expectedIcon: PiFilePngLight,
        },
        {
            name: 'works with simple file name',
            file: 'my-file.js',
            language: 'JavaScript',
            expectedIcon: SiJavascript,
        },
        {
            name: 'check unknown language',
            file: 'my-file.test.unknown',
            language: 'Unknown',
            expectedIcon: undefined,
        },
    ]

    for (const t of tests) {
        it(t.name, () => {
            const iconInfo = getFileIconInfo(t.file, t.language)
            expect(iconInfo?.icon).toBe(t.expectedIcon)
        })
    }
})
