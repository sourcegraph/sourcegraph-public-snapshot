import { testCodeHostMountGetters, testToolbarMountGetter } from '../code_intelligence/code_intelligence_test_utils'
import { bitbucketServerCodeHost, getToolbarMount } from './code_intelligence'

describe('bitbucketServerCodeHost', () => {
    testCodeHostMountGetters(bitbucketServerCodeHost, {
        getCommandPaletteMount: `${__dirname}/__fixtures__/browse.html`,
        getViewContextOnSourcegraphMount: `${__dirname}/__fixtures__/browse.html`,
    })
    describe('getToolbarMount()', () => {
        testToolbarMountGetter(`${__dirname}/__fixtures__/code-views/pull-request/split/modified.html`, getToolbarMount)
    })
})
