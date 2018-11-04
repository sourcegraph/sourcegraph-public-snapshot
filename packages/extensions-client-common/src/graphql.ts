import * as GQL from './schema/graphqlschema'

export const graphQLContent = Symbol('graphQLContent')
export interface GraphQLDocument {
    [graphQLContent]: string
}

/**
 * Use this template string tag for all GraphQL queries.
 */
export const gql = (template: TemplateStringsArray, ...substitutions: any[]): GraphQLDocument => ({
    [graphQLContent]: String.raw(template, ...substitutions.map(s => s[graphQLContent] || s)),
})

/**
 * The response from a GraphQL API query.
 */
export interface QueryResult<D extends Partial<GQL.IQuery>> {
    data?: D
    errors?: GQL.IGraphQLResponseError[]
}
