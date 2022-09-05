import isAbsoluteURL from 'is-absolute-url'

import { TourLanguage, TourTaskStepType } from './types'

/**
 * Returns a new URL w/ tour state tracking query parameters. This is used to show/hide tour task info box.
 *
 * @param href original URL
 * @param stepId TourTaskStepType id
 * @example /search?q=context:global+repo:my-repo&patternType=literal + &tour=true&stepId=DiffSearch
 */
const buildURIMarkers = (href: string, stepId: string): string => {
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

/**
 * Returns tour URL state tracking query parameters from a URL search parameter that was build using "buildURIMarkers" function.
 */
export const parseURIMarkers = (searchParameters: string): { isTour: boolean; stepId: string | null } => {
    const parameters = new URLSearchParams(searchParameters)
    const isTour = parameters.has('tour')
    const stepId = parameters.get('stepId')
    return { isTour, stepId }
}

/**
 * Check if given TourTaskStepType.action has a language specific value
 */
export const isLanguageRequired = (step: TourTaskStepType): boolean => typeof step.action.value !== 'string'

/**
 * Returns a TourTaskStepType.action.value if possible, '#' otherwise.
 */
export const getTourTaskStepActionValue = (step: TourTaskStepType, language?: TourLanguage): string =>
    typeof step.action.value === 'string'
        ? buildURIMarkers(step.action.value, step.id)
        : language
        ? buildURIMarkers(step.action.value[language], step.id)
        : '#'
