import { useState } from 'react'

import * as uuid from 'uuid'

import type { useFieldAPI } from '@sourcegraph/wildcard'

import { DEFAULT_DATA_SERIES_COLOR } from '../../../constants'

import type { EditableDataSeries } from './types'

export const createDefaultEditSeries = (series?: Partial<EditableDataSeries>): EditableDataSeries => ({
    id: `runtime-series.${uuid.v4()}`,
    ...DEFAULT_EDITABLE_SERIES,
    ...series,
})

const DEFAULT_EDITABLE_SERIES = {
    valid: false,
    edit: false,
    autofocus: false,
    name: '',
    query: '',
    stroke: DEFAULT_DATA_SERIES_COLOR,
}

export interface UseEditableSeriesAPI {
    /**
     * A list of editable data series. Basically, this is just a
     * sorted/filtered list of original data series but with
     * additional logic around create/updated/delete actions for
     * series list.
     */
    series: EditableDataSeries[]

    /** Call whenever the user changes any fields (title, query, color) of series */
    changeSeries: (liveSeries: EditableDataSeries, valid: boolean, index: number) => void

    /** Call whenever the user clicks the edit series button  in series preview card. */
    editRequest: (seriesId?: string) => void

    /** Call whenever the user clicks the save series button */
    editCommit: (editedSeries: EditableDataSeries) => void

    /**
     * Call whenever the user cancel series editing by clicking cancel button
     * in series form.
     */
    cancelEdit: (seriesId: string) => void

    /**
     * Call whenever the user tries to delete series by clicking delete button
     * in series preview card.
     */
    deleteSeries: (series: string) => void
}

/**
 * Basically this is just a stateful selector function over series that simplifies work
 * with editable series and its special UX actions like delete series through preview card,
 * edit series through form, create new series through add more series button.
 */
export function useEditableSeries(series: useFieldAPI<EditableDataSeries[]>): UseEditableSeriesAPI {
    const [seriesBeforeEdit, setSeriesBeforeEdit] = useState<Record<string, EditableDataSeries>>({})

    const handleSeriesLiveChange = (liveSeries: EditableDataSeries, valid: boolean): void => {
        series.meta.setState(state => {
            const index = state.value.findIndex(series => series.id === liveSeries.id)

            if (index !== -1) {
                const newLiveSeries = { ...liveSeries, edit: true, valid }

                return { ...state, value: replace(state.value, index, newLiveSeries) }
            }

            return state
        })
    }

    const handleEditSeriesRequest = (seriesId?: string): void => {
        const seriesValue = series.meta.value
        const newEditSeries = [...seriesValue]
        const index = newEditSeries.findIndex(series => series.id === seriesId)

        if (index !== -1) {
            newEditSeries[index] = { ...seriesValue[index], edit: true, autofocus: true }

            const newSeriesID = newEditSeries[index].id

            // If user tries edit series we have to remember the value before edit
            // in case if user clicks cancel we return that initial value back
            if (newSeriesID) {
                setSeriesBeforeEdit({
                    ...seriesBeforeEdit,
                    [newSeriesID]: seriesValue[index],
                })
            }
        } else {
            newEditSeries.push(createDefaultEditSeries({ edit: true, autofocus: true }))
        }

        series.meta.setState(state => ({ ...state, value: newEditSeries }))
    }

    const handleEditSeriesCancel = (seriesId: string): void => {
        series.meta.setState(state => {
            const index = state.value.findIndex(series => series.id === seriesId)

            if (index === -1) {
                return state
            }

            const series = [...state.value]
            const seriesValueBeforeEdit = seriesId && seriesBeforeEdit[seriesId]

            // If we have series by this index that means user activated
            // cancellation of edit mode of series that already exists
            if (seriesValueBeforeEdit) {
                // in this case we have to set values of settings that we had
                // before edit happened
                series[index] = seriesValueBeforeEdit

                return { ...state, value: series }
            }

            // On other case means that user clicked cancel of new series form
            // in this case we have to remove series model entirely
            return { ...state, value: remove(series, index) }
        })
    }

    const handleEditSeriesCommit = (editedSeries: EditableDataSeries): void => {
        series.meta.setState(state => {
            const index = state.value.findIndex(series => series.id === editedSeries.id)

            if (index === -1) {
                return state
            }

            const series = state.value
            const updatedSeries = { ...editedSeries, valid: true, edit: false }
            const newSeries = replace(series, index, updatedSeries)

            return { ...state, value: newSeries }
        })
    }

    const handleRemoveSeries = (seriesId: string): void => {
        series.meta.setState(state => {
            const index = state.value.findIndex(series => series.id === seriesId)

            if (index === -1) {
                return state
            }

            const series = state.value
            const newSeries = remove(series, index)

            if (newSeries.length === 0) {
                // If user remove all series we add fallback with another oped edit series
                // just to emphasize that user has to fill in at least one series
                return { ...state, value: [createDefaultEditSeries({ edit: true })] }
            }

            return { ...state, value: newSeries }
        })
    }

    return {
        series: series.input.value,
        changeSeries: handleSeriesLiveChange,
        editRequest: handleEditSeriesRequest,
        editCommit: handleEditSeriesCommit,
        cancelEdit: handleEditSeriesCancel,
        deleteSeries: handleRemoveSeries,
    }
}

/** Helper replace element in array by index and return new array. */
function replace<Element>(list: Element[], index: number, newElement: Element): Element[] {
    return [...list.slice(0, index), newElement, ...list.slice(index + 1)]
}

/** Helper remove element from array by index. */
function remove<Element>(list: Element[], index: number): Element[] {
    return [...list.slice(0, index), ...list.slice(index + 1)]
}
