import { testCodeHostMountGetters as testMountGetters } from '../code_intelligence/code_intelligence_test_utils'
import { gitlabCodeHost } from './code_intelligence'

describe('gitlabCodeHost', () => {
    const fixturePath = `${__dirname}/__fixtures__/repository.html`
    testMountGetters(gitlabCodeHost, fixturePath)
})
