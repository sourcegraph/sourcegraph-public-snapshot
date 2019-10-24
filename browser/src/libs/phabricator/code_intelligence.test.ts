import { testToolbarMountGetter } from '../code_intelligence/code_intelligence_test_utils'
import { commitCodeView, diffCodeView } from './code_intelligence'

describe('phabricator/code_intelligence', () => {
    describe('diffCodeView', () => {
        describe('getToolbarMount()', () => {
            for (const view of ['split', 'unified']) {
                testToolbarMountGetter(
                    `${__dirname}/__fixtures__/code-views/2017.09-r1/differential/${view}.html`,
                    diffCodeView.getToolbarMount
                )
            }
        })
    })
    describe('commitCodeView', () => {
        describe('getToolbarMount()', () => {
            for (const view of ['split', 'unified']) {
                testToolbarMountGetter(
                    `${__dirname}/__fixtures__/code-views/2017.09-r1/commit/${view}.html`,
                    commitCodeView.getToolbarMount
                )
            }
        })
    })
    // TODO sourceCodeView, currently not possible because code view element does not contain toolbar
})
