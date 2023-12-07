export enum Param {
    before = '$before',
    after = '$after',
    last = '$last',
}

export function getPaginationParams(
    searchParams: URLSearchParams,
    pageSize: number
):
| { first: number; last: null; before: null; after: string | null }
| { first: null; last: number; before: string | null; after: null } {
    if (searchParams.has('$before')) {
        return { first: null, last: pageSize, before: searchParams.get(Param.before), after: null }
    }
    if (searchParams.has('$after')) {
        return { first: pageSize, last: null, before: null, after: searchParams.get(Param.after) }
    }
    if (searchParams.has('$last')) {
        return { first: null, last: pageSize, before: null, after: null }
    }
    return { first: pageSize, last: null, before: null, after: null }
}
