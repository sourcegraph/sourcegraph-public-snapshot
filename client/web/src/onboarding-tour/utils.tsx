import isAbsoluteURL from 'is-absolute-url'

export const buildURIMarkers = (href: string, stepId: string): string => {
    const isRelative = !isAbsoluteURL(href)

    try {
        const url = new URL(href, isRelative ? `${location.protocol}//${location.host}` : undefined)
        url.searchParams.set('tour', 'true')
        url.searchParams.set('stepId', stepId)
        return isRelative ? url.toString().slice(url.origin.length) : url.toString()
    } catch {
        return '#'
    }
}

export const parseURIMarkers = (searchParameters: string): { isTour: boolean; stepId: string | null } => {
    const parameters = new URLSearchParams(searchParameters)
    const isTour = parameters.has('tour')
    const stepId = parameters.get('stepId')
    return { isTour, stepId }
}

export const isExternalURL = (url: string): boolean => isAbsoluteURL(url) && new URL(url).origin !== location.origin
