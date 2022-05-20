import { Dispatch, SetStateAction, useState } from 'react'

interface UseSeriesToggleReturn {
    // Currently hovered series id
    hoveredId: string | undefined

    // List of currently selected series ids
    selectedSeriesIds: string[]

    // These functions are exposed to keep all of the state
    // contained in the hook
    isSelected: (id: string) => boolean
    setHoveredId: Dispatch<SetStateAction<string | undefined>>
    toggle: (id: string) => void
}

// Used when clicking legend items in a line chart. This hook manages the currently
// selected data series as well as the currently hovered legend item.
export const useSeriesToggle = (): UseSeriesToggleReturn => {
    const [selectedSeriesIds, setSelectedSeriesIds] = useState<string[]>([])
    const [hoveredId, setHoveredId] = useState<string | undefined>()

    const selectSeries = (seriesId: string): void => setSelectedSeriesIds([...selectedSeriesIds, seriesId])
    const deselectSeries = (seriesId: string): void =>
        setSelectedSeriesIds(selectedSeriesIds.filter(id => id !== seriesId))
    const toggle = (seriesId: string): void =>
        selectedSeriesIds.includes(seriesId) ? deselectSeries(seriesId) : selectSeries(seriesId)
    const isSelected = (seriesId: string): boolean => {
        if (selectedSeriesIds.length === 0) {
            return true
        }

        return selectedSeriesIds.includes(seriesId)
    }

    return {
        // state
        hoveredId,
        selectedSeriesIds,

        // functions
        isSelected,
        setHoveredId,
        toggle,
    }
}
