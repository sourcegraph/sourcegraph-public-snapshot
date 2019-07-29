import * as GQL from '../../../../shared/src/graphql/schema'

export function threadNoun(type: GQL.ThreadType, plural = false): string {
    return `${type.toLowerCase()}${plural ? 's' : ''}`
}
