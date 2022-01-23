import { Location, LocationDescriptorObject } from 'history'
import { useCallback, useMemo } from 'react'
import { useHistory, useLocation } from 'react-router'

export type TreeOrComponentViewMode = 'auto' | 'tree'

export interface TreeOrComponentViewOptionsProps {
    treeOrComponentViewMode: TreeOrComponentViewMode
    treeOrComponentViewModeURL: Record<TreeOrComponentViewMode, LocationDescriptorObject>
    setTreeOrComponentViewMode: (mode: TreeOrComponentViewMode) => void
}

export function useTreeOrComponentViewOptions(): TreeOrComponentViewOptionsProps {
    const location = useLocation()
    const history = useHistory()

    const treeOrComponentViewMode: TreeOrComponentViewMode = useMemo(
        () => (new URLSearchParams(location.search).get('as') === 'tree' ? 'tree' : 'auto'),
        [location.search]
    )
    const treeOrComponentViewModeURL = useMemo<TreeOrComponentViewOptionsProps['treeOrComponentViewModeURL']>(
        () => ({
            auto: makeTreeOrComponentViewURL(location, 'auto'),
            tree: makeTreeOrComponentViewURL(location, 'tree'),
        }),
        [location]
    )
    const setTreeOrComponentViewMode = useCallback<TreeOrComponentViewOptionsProps['setTreeOrComponentViewMode']>(
        mode => history.push(makeTreeOrComponentViewURL(location, mode)),
        [history, location]
    )

    return useMemo(() => ({ treeOrComponentViewMode, treeOrComponentViewModeURL, setTreeOrComponentViewMode }), [
        setTreeOrComponentViewMode,
        treeOrComponentViewMode,
        treeOrComponentViewModeURL,
    ])
}

function makeTreeOrComponentViewURL(location: Location, mode: TreeOrComponentViewMode): LocationDescriptorObject {
    const search = new URLSearchParams(location.search)
    if (mode === 'tree') {
        search.set('as', mode)
    } else {
        search.delete('as')
    }

    return { ...location, search: search.toString() }
}
