import { Range } from '@sourcegraph/extension-api-types'

import { LocationFields } from '../graphql-operations'

export interface Location {
    repo: string
    file: string
    content: string
    commitID: string
    range?: Range
    url: string
    lines: string[]
    precise: boolean
}

export interface RepoLocationGroup {
    repoName: string
    referenceGroups: LocationGroup[]
}

export interface LocationGroup {
    repoName: string
    path: string
    locations: Location[]
}

export const buildPreciseLocation = (node: LocationFields): Location => buildLocation(node, true)
export const buildSearchBasedLocation = (node: LocationFields): Location => buildLocation(node, false)

const buildLocation = (node: LocationFields, precise: boolean): Location => {
    const location: Location = {
        content: node.resource.content,
        commitID: node.resource.commit.oid,
        repo: node.resource.repository.name,
        file: node.resource.path,
        url: '',
        lines: [],
        precise,
    }
    if (node.range !== null) {
        location.range = node.range
    }
    location.url = node.url
    location.lines = location.content.split(/\r?\n/)
    return location
}

export type LocationGroupQuality = 'PRECISE' | 'SEARCH-BASED' | 'MIXED'

export const locationGroupQuality = (locations: Location[]): LocationGroupQuality => {
    let hasPrecise = false
    let hasSearchBased = false

    for (const location of locations) {
        if (location.precise) {
            hasPrecise = true
        } else {
            hasSearchBased = true
        }
    }

    if (hasPrecise && hasSearchBased) {
        return 'MIXED'
    }
    if (hasPrecise) {
        return 'PRECISE'
    }
    return 'SEARCH-BASED'
}
