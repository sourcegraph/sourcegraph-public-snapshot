import { existsSync, readdirSync } from 'fs'
import { startCase } from 'lodash'
import { testCodeHostMountGetters, testToolbarMountGetter } from '../code_intelligence/code_intelligence_test_utils'
import { CodeView } from '../code_intelligence/code_views'
import { createFileActionsToolbarMount, createFileLineContainerToolbarMount, githubCodeHost } from './code_intelligence'

const testCodeHost = (fixturePath: string) => {
    if (existsSync(fixturePath)) {
        describe('githubCodeHost', () => {
            testCodeHostMountGetters(githubCodeHost, fixturePath)
        })
    }
}

const testMountGetter = (
    mountGetter: NonNullable<CodeView['getToolbarMount']>,
    mountGetterName: string,
    codeViewFixturePath: string
) => {
    if (existsSync(codeViewFixturePath)) {
        describe(mountGetterName, () => {
            testToolbarMountGetter(codeViewFixturePath, mountGetter)
        })
    }
}

describe('github/code_intelligence', () => {
    for (const version of ['github.com', 'ghe-2.14.11']) {
        describe(version, () => {
            for (const page of readdirSync(`${__dirname}/__fixtures__/${version}`)) {
                describe(`${startCase(page)} page`, () => {
                    for (const extension of ['vanilla', 'refined-github']) {
                        describe(startCase(extension), () => {
                            if (page === 'blob') {
                                // no split/unified view on blobs
                                const directory = `${__dirname}/__fixtures__/${version}/${page}/${extension}`
                                testCodeHost(`${directory}/page.html`)
                                testMountGetter(
                                    createFileLineContainerToolbarMount,
                                    'createSingleFileToolbarMount()',
                                    `${directory}/code-view.html`
                                )
                            }
                            for (const view of ['split', 'unified']) {
                                describe(`${startCase(view)} view`, () => {
                                    const directory = `${__dirname}/__fixtures__/${version}/${page}/${extension}/${view}`
                                    testCodeHost(`${directory}/page.html`)
                                    describe('createCodeViewToolbarMount()', () => {
                                        testMountGetter(
                                            createFileActionsToolbarMount,
                                            'createCodeViewToolbarMount()',
                                            `${directory}/code-view.html`
                                        )
                                    })
                                })
                            }
                        })
                    }
                })
            }
        })
    }
})
