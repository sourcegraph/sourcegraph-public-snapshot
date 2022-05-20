import { Dispatch, SetStateAction, useState } from 'react'

interface UseSeriesToggleReturn {
    toggle: (id: string) => void
    selectedSeriesIds: string[]
    isSelected: (id: string) => boolean
    hoveredId: string | undefined
    setHoveredId: Dispatch<SetStateAction<string | undefined>>
}

export const useSeriesToggle = (currentSelectedSeriesIds: string[]): UseSeriesToggleReturn => {
    const [selectedSeriesIds, setSelectedSeriesIds] = useState<string[]>(currentSelectedSeriesIds)
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
        toggle,
        selectedSeriesIds,
        isSelected,
        hoveredId,
        setHoveredId,
    }
}
