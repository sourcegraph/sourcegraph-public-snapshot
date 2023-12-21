import { groupBy } from 'lodash'

import { CommitMatch, Filter, SearchMatch } from '@sourcegraph/shared/src/search/stream'

export const generateAuthorFilters = (results: SearchMatch[]): Filter[] => {
    const commitMatches = results.filter(match => match.type === 'commit') as CommitMatch[]
    const groupedMatches = groupBy(commitMatches, match => match.authorName)

    return Object.keys(groupedMatches)
        .map<Filter>(authorName => {
            const commitMatches = groupedMatches[authorName]

            return {
                kind: 'author',
                label: authorName,
                count: commitMatches.length,
                value: `author:"${authorName}"`,
                limitHit: false,
            }
        })
        .sort((match1, match2) => match2.count - match1.count)
}
