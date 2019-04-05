import { testMountGetterInvariants } from '../code_intelligence/code_intelligence_test_utils'
import { bitbucketServerCodeHost } from './code_intelligence'

describe('bitbucketServerCodeHost', () => {
    const fixturePath = `${__dirname}/__fixtures__/pr-modified.html`
    testMountGetterInvariants(bitbucketServerCodeHost, {
        getCommandPaletteMount: fixturePath,
        getViewContextOnSourcegraphMount: `${__dirname}/__fixtures__/browse.html`,
    })
})
