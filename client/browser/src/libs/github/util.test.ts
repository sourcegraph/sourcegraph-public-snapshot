import { startCase } from 'lodash'
import { readFile } from 'mz/fs'
import { isDomSplitDiff, parseGitHubHash } from './util'

describe('util', () => {
    describe('parseGitHubHash()', () => {
        it('parses nonexistent', () => expect(parseGitHubHash('')).toBe(undefined))
        it('parses empty', () => expect(parseGitHubHash('#')).toBe(undefined))
        it('parses single line', () => expect(parseGitHubHash('#L123')).toEqual({ startLine: 123, endLine: undefined }))
        it('parses range', () => expect(parseGitHubHash('#L123-L456')).toEqual({ startLine: 123, endLine: 456 }))
        it('handles invalid value', () => expect(parseGitHubHash('#Lfoo')).toBe(undefined))
        it('allows extra after', () =>
            expect(parseGitHubHash('#L123-L456-foo')).toEqual({ startLine: 123, endLine: 456 }))
    })

    describe('isDomSplitDiff()', () => {
        for (const version of ['github.com', 'ghe-2.14.11']) {
            describe(`Version ${version}`, () => {
                const views = [
                    {
                        view: 'pull-request',
                        url: 'https://github.com/sourcegraph/sourcegraph/pull/2672/files',
                    },
                    {
                        view: 'commit',
                        url:
                            'https://github.com/sourcegraph/sourcegraph/commit/2c74f329fd03008fa0b446cd5e53234715dae3dc',
                    },
                ]
                for (const { view, url } of views) {
                    describe(`${startCase(view)} page`, () => {
                        beforeEach(() => {
                            jsdom.reconfigure({ url })
                        })
                        for (const extension of ['vanilla', 'refined-github']) {
                            describe(startCase(extension), () => {
                                it('should return true for split view', async () => {
                                    document.body.innerHTML = await readFile(
                                        `${__dirname}/__fixtures__/${version}/${view}/${extension}/split.html`,
                                        'utf-8'
                                    )
                                    expect(isDomSplitDiff()).toBe(true)
                                })
                                it('should return false for unified view', async () => {
                                    document.body.innerHTML = await readFile(
                                        `${__dirname}/__fixtures__/${version}/${view}/${extension}/unified.html`,
                                        'utf-8'
                                    )
                                    expect(isDomSplitDiff()).toBe(false)
                                })
                            })
                        }
                    })
                }
            })
        }
    })
})
