import { Scalars } from '@sourcegraph/shared/src/graphql-operations'
import { hasProperty } from '@sourcegraph/shared/src/util/types'

/** Checks if the passed value satisfies the GraphQL Node interface */
export const hasID = (value: unknown): value is { id: Scalars['ID'] } =>
    typeof value === 'object' && value !== null && hasProperty('id')(value) && typeof value.id === 'string'
