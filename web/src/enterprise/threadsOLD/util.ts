import * as GQL from '../../../../shared/src/graphql/schema'

export function threadNoun(type: any, plural = false): string {
    return `${type.toLowerCase()}${plural ? 's' : ''}`
}
