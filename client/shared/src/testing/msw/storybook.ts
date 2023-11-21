import type { TypeMocks } from '../../graphql-types'
import { defaultMocks } from '../graphql/defaultMocks'

import { type MockHandler, setupHandlers as defaultSetupHandlers } from './handler'

const schemas: Record<string, string> = import.meta.glob('../../../../../cmd/frontend/graphqlbackend/*.graphql', {
    as: 'raw',
    eager: true,
})

export function setupHandlers(): MockHandler<TypeMocks> {
    return defaultSetupHandlers({ typeDefs: Object.values(schemas).join('\n'), defaultMocks })
}
