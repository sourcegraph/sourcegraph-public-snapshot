import { startCase } from 'lodash'
import { testMountGetterInvariants } from '../code_intelligence/code_intelligence_test_utils'
import { githubCodeHost } from './code_intelligence'

describe('githubCodeHost', () => {
    for (const version of ['github.com', 'ghe-2.14.11']) {
        describe(version, () => {
            for (const page of ['commit', 'pull-request']) {
                describe(`${startCase(page)} page`, () => {
                    for (const extension of ['vanilla', 'refined-github']) {
                        describe(startCase(extension), () => {
                            for (const view of ['split', 'unified']) {
                                describe(`${view} view`, () => {
                                    const fixturePath = `${__dirname}/__fixtures__/${version}/${page}/${extension}/${view}.html`
                                    testMountGetterInvariants(githubCodeHost, fixturePath)
                                })
                            }
                        })
                    }
                })
            }
        })
    }
})
