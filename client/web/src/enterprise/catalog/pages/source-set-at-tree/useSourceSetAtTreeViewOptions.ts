import { Location, LocationDescriptorObject } from 'history'
import { useCallback, useMemo } from 'react'
import { useHistory, useLocation } from 'react-router'

export type SourceSetAtTreeViewMode = 'auto' | 'tree'

export interface SourceSetAtTreeViewOptionsProps {
    sourceSetAtTreeViewMode: SourceSetAtTreeViewMode
    sourceSetAtTreeViewModeURL: Record<SourceSetAtTreeViewMode, LocationDescriptorObject>
    setSourceSetAtTreeViewMode: (mode: SourceSetAtTreeViewMode) => void
}

export function useSourceSetAtTreeViewOptions(): SourceSetAtTreeViewOptionsProps {
    const location = useLocation()
    const history = useHistory()

    const sourceSetAtTreeViewMode: SourceSetAtTreeViewMode = useMemo(
        () => (new URLSearchParams(location.search).get('as') === 'tree' ? 'tree' : 'auto'),
        [location.search]
    )
    const sourceSetAtTreeViewModeURL = useMemo<SourceSetAtTreeViewOptionsProps['sourceSetAtTreeViewModeURL']>(
        () => ({
            auto: makeSourceSetAtTreeViewURL(location, 'auto'),
            tree: makeSourceSetAtTreeViewURL(location, 'tree'),
        }),
        [location]
    )
    const setSourceSetAtTreeViewMode = useCallback<SourceSetAtTreeViewOptionsProps['setSourceSetAtTreeViewMode']>(
        mode => history.push(makeSourceSetAtTreeViewURL(location, mode)),
        [history, location]
    )

    return useMemo(() => ({ sourceSetAtTreeViewMode, sourceSetAtTreeViewModeURL, setSourceSetAtTreeViewMode }), [
        setSourceSetAtTreeViewMode,
        sourceSetAtTreeViewMode,
        sourceSetAtTreeViewModeURL,
    ])
}

function makeSourceSetAtTreeViewURL(location: Location, mode: SourceSetAtTreeViewMode): LocationDescriptorObject {
    const search = new URLSearchParams(location.search)
    if (mode === 'tree') {
        search.set('as', mode)
    } else {
        search.delete('as')
    }

    return { ...location, search: search.toString() }
}
