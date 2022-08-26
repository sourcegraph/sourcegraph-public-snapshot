import mock from 'mock-require'

console.log('BOOM')
// Stub Sourcegraph API
import { createStubSourcegraphAPI } from '@sourcegraph/extension-api-stubs'

mock('sourcegraph', createStubSourcegraphAPI())

test('init.placeholder', () => {})

export const placeholder = 42
