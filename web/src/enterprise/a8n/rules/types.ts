import * as GQL from '../../../../../shared/src/graphql/schema'

export type RuleScope = Pick<GQL.IProject, '__typename' | 'id'>
