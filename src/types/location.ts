import { Range } from './range'

export interface Location {
    uri: string
    range: Range
}

/**
 * The definition of a symbol represented as one or many [locations](#Location). For most programming languages
 * there is only one location at which a symbol is defined. If no definition can be found `null` is returned.
 */
export type Definition = Location | Location[] | null
