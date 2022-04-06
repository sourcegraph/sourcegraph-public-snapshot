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

export type LocationGroupQuality = 'PRECISE' | 'SEARCH-BASED'

export const locationGroupQuality = (group: LocationGroup): LocationGroupQuality => {
    if (group.locations.length === 0) {
        throw new Error(`No locations in ${group.path}`)
    }

    // Since we don't mix precise & search-based in a single file, we know that in a
    // single group all locations are either search-based or precise.
    return group.locations[0].precise ? 'PRECISE' : 'SEARCH-BASED'
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

export const buildRepoLocationGroups = (locations: Location[]): RepoLocationGroup[] => {
    const byRepoAndFile: Record<string, Record<string, Location[]>> = {}
    for (const location of locations) {
        if (byRepoAndFile[location.repo] === undefined) {
            byRepoAndFile[location.repo] = {}
        }
        if (byRepoAndFile[location.repo][location.file] === undefined) {
            byRepoAndFile[location.repo][location.file] = []
        }
        byRepoAndFile[location.repo][location.file].push(location)
    }

    return Object.keys(byRepoAndFile).map(repoName => {
        const byFile = byRepoAndFile[repoName]

        const referenceGroups: LocationGroup[] = Object.keys(byFile).map(path => ({
            path,
            locations: byFile[path],
            repoName,
        }))

        return { repoName, referenceGroups }
    })
}
