import { setupHandlers as defaultSetupHandlers } from './handler'

const schemas: Record<string, string> = import.meta.glob('../../../../../cmd/frontend/graphqlbackend/*.graphql', {
    as: 'raw',
    eager: true,
})

export function setupHandlers(): ReturnType<typeof defaultSetupHandlers> {
    return defaultSetupHandlers({ typeDefs: Object.values(schemas).join('\n') })
}
