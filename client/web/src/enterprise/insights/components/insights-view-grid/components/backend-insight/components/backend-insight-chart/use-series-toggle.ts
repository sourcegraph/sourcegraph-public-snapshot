import { useState } from 'react'

type SeriesId = string | number
type SeriesIds = SeriesId[]

interface UseSeriesToggleReturn {
    toggle: (id: SeriesId) => void
    selectedSeriesIds: SeriesIds
    isSelected: (id: SeriesId) => boolean
}

export const useSeriesToggle = (currentSelectedSeriesIds: SeriesIds): UseSeriesToggleReturn => {
    const [selectedSeriesIds, setSelectedSeriesIds] = useState<SeriesIds>(currentSelectedSeriesIds)

    const selectSeries = (seriesId: SeriesId): void => setSelectedSeriesIds([...selectedSeriesIds, seriesId])
    const deselectSeries = (seriesId: SeriesId): void =>
        setSelectedSeriesIds(selectedSeriesIds.filter(id => id !== seriesId))
    const toggle = (seriesId: SeriesId): void =>
        selectedSeriesIds.includes(seriesId) ? deselectSeries(seriesId) : selectSeries(seriesId)
    const isSelected = (seriesId: SeriesId): boolean => {
        if (selectedSeriesIds.length === 0) {
            return true
        }

        return selectedSeriesIds.includes(seriesId)
    }

    return {
        toggle,
        selectedSeriesIds,
        isSelected,
    }
}
