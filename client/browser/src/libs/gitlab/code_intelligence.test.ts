import {
    testCodeHostMountGetters as testMountGetters,
    testToolbarMountGetter,
} from '../code_intelligence/code_intelligence_test_utils'
import { getToolbarMount, gitlabCodeHost } from './code_intelligence'

describe('gitlab/code_intelligence', () => {
    describe('gitlabCodeHost', () => {
        testMountGetters(gitlabCodeHost, `${__dirname}/__fixtures__/repository.html`)
    })
    describe('getToolbarMount()', () => {
        testToolbarMountGetter(`${__dirname}/__fixtures__/code-views/pr-unified.html`, getToolbarMount)
    })
})
