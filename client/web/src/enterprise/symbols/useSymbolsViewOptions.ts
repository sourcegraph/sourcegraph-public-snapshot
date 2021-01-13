import H from 'history'
import { useMemo } from 'react'
import * as GQL from '../../../../shared/src/graphql/schema'

export interface SymbolsViewOptionsProps {
    viewOptions: SymbolsViewOptions
}

export interface SymbolsViewOptions {
    internals: GQL.ISymbolFilters['internals']
    externals: boolean
}

const DEFAULT_OPTIONS: SymbolsViewOptions = {
    externals: true,
    internals: false,
}

const KEYS: (keyof SymbolsViewOptions)[] = ['externals', 'internals']

interface ToggleURLs {
    externals: H.LocationDescriptorObject
    internals: H.LocationDescriptorObject
}

interface Props {
    location: H.Location
}

const locationWithViewOptions = (
    base: H.LocationDescriptorObject,
    viewOptions: SymbolsViewOptions
): H.LocationDescriptorObject => {
    const parameters = new URLSearchParams(base.search)

    for (const key of KEYS) {
        if (viewOptions[key] === DEFAULT_OPTIONS[key]) {
            parameters.delete(key)
        } else {
            parameters.set(key, viewOptions[key] ? '1' : '0')
        }
    }

    return { ...base, search: parameters.toString() }
}

const parseSearchParameterValue = (value: string | null, defaultValue: boolean): boolean =>
    value === null ? defaultValue : value === '1'

export const useSymbolsViewOptions = ({
    location,
}: Props): { viewOptions: SymbolsViewOptions; toggleURLs: ToggleURLs } => {
    const viewOptions = useMemo<SymbolsViewOptions>(() => {
        const parameters = new URLSearchParams(location.search)
        return {
            externals: parseSearchParameterValue(parameters.get('externals'), DEFAULT_OPTIONS.externals),
            internals: parseSearchParameterValue(parameters.get('internals'), DEFAULT_OPTIONS.internals),
        }
    }, [location.search])

    const toggleURLs = useMemo<ToggleURLs>(
        () => ({
            externals: locationWithViewOptions(location, { ...viewOptions, externals: !viewOptions.externals }),
            internals: locationWithViewOptions(location, { ...viewOptions, internals: !viewOptions.internals }),
        }),
        [location, viewOptions]
    )

    return { viewOptions, toggleURLs }
}
