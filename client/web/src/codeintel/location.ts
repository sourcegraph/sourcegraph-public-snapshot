import { Range } from '@sourcegraph/extension-api-types'

import { LocationFields } from '../graphql-operations'

export interface Location {
    resource: {
        path: string
        content: string
        repository: {
            name: string
        }
        commit: {
            oid: string
        }
    }
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
        resource: {
            repository: { name: node.resource.repository.name },
            content: node.resource.content,
            path: node.resource.path,
            commit: node.resource.commit,
        },
        url: '',
        lines: [],
        precise,
    }
    if (node.range !== null) {
        location.range = node.range
    }
    location.url = node.url
    location.lines = location.resource.content.split(/\r?\n/)
    return location
}
