import { testToolbarMountGetter } from '../code_intelligence/code_intelligence_test_utils'
import { diffCodeView } from './code_intelligence'

describe('phabricator/code_intelligence', () => {
    describe('diffCodeView', () => {
        describe('getToolbarMount()', () => {
            testToolbarMountGetter(
                `${__dirname}/__fixtures__/code-views/diff-side-by-side.html`,
                diffCodeView.getToolbarMount
            )
        })
    })
    // TODO sourceCodeView
    // TODO commitCodeView
})
