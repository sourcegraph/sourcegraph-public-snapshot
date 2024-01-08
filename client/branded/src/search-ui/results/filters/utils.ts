import { groupBy } from 'lodash'

import { CommitMatch, SearchMatch } from '@sourcegraph/shared/src/search/stream'

import { DynamicClientFilter } from './types'

export const generateAuthorFilters = (results: SearchMatch[]): DynamicClientFilter[] => {
    const commitMatches = results.filter(match => match.type === 'commit') as CommitMatch[]
    const groupedMatches = groupBy(commitMatches, match => match.authorName)

    return Object.keys(groupedMatches)
        .map<DynamicClientFilter>(authorName => {
            const commitMatches = groupedMatches[authorName]

            return {
                kind: 'author',
                label: authorName,
                count: commitMatches.length,
                value: `author:"${authorName}"`,
                exhaustive: true,
            }
        })
        .sort((match1, match2) => match2.count - match1.count)
}
