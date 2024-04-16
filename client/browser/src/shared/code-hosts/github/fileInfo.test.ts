import { readFile } from 'mz/fs'
import { describe, expect, it } from 'vitest'

import { getFilePath, getFilePathFromURL } from './util'

const tests = [
    ['github.com/blob/vanilla/page.html', 'shared/src/api/extension/types/url.ts'],
    ['github.com/blob/refined-github/page.html', 'shared/src/api/extension/types/url.ts'],
    ['ghe-2.14.11/blob/vanilla/page.html', 'bench_test.go'],
    ['ghe-2.14.11/blob/refined-github/page.html', 'bench_test.go'],
]

describe('github/fileInfo', () => {
    describe('getFilePath()', () => {
        for (const [fixture, expectedFilePath] of tests) {
            it(`finds the file path in ${fixture}`, async () => {
                document.body.innerHTML = await readFile(`${__dirname}/__fixtures__/${fixture}`, 'utf-8')
                expect(getFilePath()).toBe(expectedFilePath)
            })
        }
    })

    describe('getFilePathFromURL()', () => {
        const toReturn = [
            {
                url: 'https://github.com/sourcegraph/sourcegraph/blob/4.4/client/browser/src/browser-extension/browser-action-icon.ts',
                rev: '4.4',
                filePath: 'client/browser/src/browser-extension/browser-action-icon.ts',
            },
            {
                url: 'https://github.com/sourcegraph/sourcegraph/blob/bext/release/client/browser/src/browser-extension/browser-action-icon.ts',
                rev: 'bext/release',
                filePath: 'client/browser/src/browser-extension/browser-action-icon.ts',
            },
            {
                url: 'https://github.com/sourcegraph/sourcegraph/blob/sourcegraph/sourcegraph/client/browser/src/browser-extension/browser-action-icon.ts',
                rev: 'sourcegraph/sourcegraph',
                filePath: 'client/browser/src/browser-extension/browser-action-icon.ts',
            },
        ]

        for (const { url, rev, filePath } of toReturn) {
            it(`returns "${filePath}" for URL "${url}" and revision "${rev}"`, () => {
                expect(getFilePathFromURL(rev, new URL(url))).toBe(filePath)
            })
        }

        const toThrow = [
            {
                url: 'https://github.com/sourcegraph/sourcegraph/blob/4.4',
                rev: '4.4',
                reason: 'no file in blob page URL',
            },
            {
                url: 'https://github.com/sourcegraph/sourcegraph/blob/main/client/browser/src/browser-extension/browser-action-icon.ts',
                rev: 'bext/release',
                reason: 'revision does not match',
            },
        ]

        for (const { url, rev, reason } of toThrow) {
            it(`throws an error for URL "${url}" and revision "${rev}", reason: "${reason}"`, () => {
                expect(() => {
                    getFilePathFromURL(rev, new URL(url))
                }).toThrow()
            })
        }
    })
})
