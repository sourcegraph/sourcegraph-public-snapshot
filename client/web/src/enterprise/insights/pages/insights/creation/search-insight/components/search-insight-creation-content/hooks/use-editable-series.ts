import { useState } from 'react'

import * as uuid from 'uuid'

import { useFieldAPI } from '../../../../../../../components/form/hooks/useField'
import { DEFAULT_DATA_SERIES_COLOR } from '../../../constants'
import { CreateInsightFormFields, EditableDataSeries } from '../../../types'

import { remove, replace } from './helpers'

const EDIT_SERIES_PREFIX = 'runtime-series'

export const createDefaultEditSeries = (series?: Partial<EditableDataSeries>): EditableDataSeries => ({
    id: `${EDIT_SERIES_PREFIX}.${uuid.v4()}`,
    ...defaultEditSeries,
    ...series,
})

const defaultEditSeries = {
    valid: false,
    edit: false,
    name: '',
    query: '',
    stroke: DEFAULT_DATA_SERIES_COLOR,
}

export interface UseEditableSeriesProps {
    series: useFieldAPI<CreateInsightFormFields['series']>
}

export interface UseEditableSeriesAPI {
    /**
     * Edit series array used below for rendering series edit form.
     * In case of some element has undefined value we're showing
     * series card with data instead of form.
     * */
    editSeries: CreateInsightFormFields['series']

    /**
     * Handler to listen latest values of particular sereis form.
     * */
    listen: (liveSeries: EditableDataSeries, valid: boolean, index: number) => void

    /**
     * Handlers for CRUD operations over series.
     * */
    editRequest: (seriesId?: string) => void
    editCommit: (editedSeries: EditableDataSeries) => void
    cancelEdit: (seriesId: string) => void
    deleteSeries: (series: string) => void
}

/**
 * Implementation of CRUD operation over insight series. Used in form to manage
 * edit, delete, add, and cancel series forms.
 */
export function useEditableSeries(props: UseEditableSeriesProps): UseEditableSeriesAPI {
    const { series } = props

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
            newEditSeries[index] = { ...seriesValue[index], edit: true }

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
            newEditSeries.push(createDefaultEditSeries({ edit: true }))
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
        editSeries: series.input.value,
        listen: handleSeriesLiveChange,
        editRequest: handleEditSeriesRequest,
        editCommit: handleEditSeriesCommit,
        cancelEdit: handleEditSeriesCancel,
        deleteSeries: handleRemoveSeries,
    }
}
