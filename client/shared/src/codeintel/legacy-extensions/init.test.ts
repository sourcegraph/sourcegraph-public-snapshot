import mock from 'mock-require'

// Stub Sourcegraph API
import { createStubSourcegraphAPI } from '@sourcegraph/extension-api-stubs'

mock('sourcegraph', createStubSourcegraphAPI())

describe('init', () => {
    it('placeholder', () => {})
})
