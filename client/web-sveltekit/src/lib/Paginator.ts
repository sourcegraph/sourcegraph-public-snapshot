export enum Param {
    before = '_before',
    after = '_after',
    last = '_last',
}

export function getPaginationParams(
    searchParams: URLSearchParams,
    pageSize: number
):
    | { first: number; last: null; before: null; after: string | null }
    | { first: null; last: number; before: string | null; after: null } {
    if (searchParams.has(Param.before)) {
        return { first: null, last: pageSize, before: searchParams.get(Param.before), after: null }
    }
    if (searchParams.has(Param.after)) {
        return { first: pageSize, last: null, before: null, after: searchParams.get(Param.after) }
    }
    if (searchParams.has(Param.last)) {
        return { first: null, last: pageSize, before: null, after: null }
    }
    return { first: pageSize, last: null, before: null, after: null }
}
