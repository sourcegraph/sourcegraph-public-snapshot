import mock from 'mock-require'
import { describe, it } from 'vitest'

// Stub Sourcegraph API
import { createStubSourcegraphAPI } from '@sourcegraph/extension-api-stubs'

mock('sourcegraph', createStubSourcegraphAPI())

describe('init', () => {
    it('placeholder', () => {})
})
