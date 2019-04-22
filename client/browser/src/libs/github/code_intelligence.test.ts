import { existsSync, readdirSync } from 'fs'
import { startCase } from 'lodash'
import { testCodeHostMountGetters, testToolbarMountGetter } from '../code_intelligence/code_intelligence_test_utils'
import { createCodeViewToolbarMount, githubCodeHost } from './code_intelligence'

describe('github/code_intelligence', () => {
    for (const version of ['github.com', 'ghe-2.14.11']) {
        describe(version, () => {
            for (const page of readdirSync(`${__dirname}/__fixtures__/${version}`)) {
                describe(`${startCase(page)} page`, () => {
                    for (const extension of ['vanilla', 'refined-github']) {
                        describe(startCase(extension), () => {
                            for (const view of ['split', 'unified']) {
                                describe(`${startCase(view)} view`, () => {
                                    const directory = `${__dirname}/__fixtures__/${version}/${page}/${extension}/${view}`
                                    const fixturePath = `${directory}/page.html`
                                    if (existsSync(fixturePath)) {
                                        describe('githubCodeHost', () => {
                                            testCodeHostMountGetters(githubCodeHost, fixturePath)
                                        })
                                    }
                                    const codeViewFixturePath = `${directory}/code-view.html`
                                    if (existsSync(codeViewFixturePath)) {
                                        describe('createCodeViewToolbarMount()', () => {
                                            testToolbarMountGetter(codeViewFixturePath, createCodeViewToolbarMount)
                                        })
                                    }
                                    // TODO test fileLineContainerCodeView.getToolbarMount()
                                })
                            }
                        })
                    }
                })
            }
        })
    }
})
