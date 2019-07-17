import * as GQL from '../../../../../shared/src/graphql/schema'

// TODO!(sqs): replace with GQL.IRule#url
export const urlToRule = (rulesURL: string, rule: Pick<GQL.IRule, 'id'>): string => `${rulesURL}/${rule.id}`
