import { useState } from 'react'
import * as uuid from 'uuid'

import { useFieldAPI } from '../../../../../../components/form/hooks/useField'
import { CreateInsightFormFields, EditableDataSeries } from '../../../types'
import { DEFAULT_ACTIVE_COLOR } from '../../form-color-input/FormColorInput'

import { remove, replace } from './helpers'

export const createDefaultEditSeries = (series?: Partial<EditableDataSeries>): EditableDataSeries => ({
    id: uuid.v4(),
    ...defaultEditSeries,
    ...series,
})

const defaultEditSeries = {
    valid: false,
    edit: false,
    name: '',
    query: '',
    stroke: DEFAULT_ACTIVE_COLOR,
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
    editRequest: (index: number) => void
    editCommit: (index: number, editedSeries: EditableDataSeries) => void
    cancelEdit: (index: number) => void
    deleteSeries: (index: number) => void
}

/**
 * Implementation of CRUD operation over insight series. Used in form to manage
 * edit, delete, add, and cancel series forms.
 * */
export function useEditableSeries(props: UseEditableSeriesProps): UseEditableSeriesAPI {
    const { series } = props

    const [seriesBeforeEdit, setSeriesBeforeEdit] = useState<Record<string, EditableDataSeries>>({})

    const handleSeriesLiveChange = (liveSeries: EditableDataSeries, valid: boolean, index: number): void => {
        series.meta.setState(state => {
            const newLiveSeries = { ...liveSeries, edit: true, valid }

            return { ...state, value: replace(state.value, index, newLiveSeries) }
        })
    }

    const handleEditSeriesRequest = (index: number): void => {
        const seriesValue = series.meta.value
        const newEditSeries = [...seriesValue]

        newEditSeries[index] = seriesValue[index]
            ? { ...seriesValue[index], edit: true }
            : createDefaultEditSeries({ edit: true })

        // If user tries edit series we have to remember value before edit
        // in case if user clicks cancel we return that initial value back
        if (seriesValue[index]) {
            const newSeriesID = newEditSeries[index].id

            setSeriesBeforeEdit({
                ...seriesBeforeEdit,
                [newSeriesID]: seriesValue[index],
            })
        }

        series.meta.setState(state => ({ ...state, value: newEditSeries }))
    }

    const handleEditSeriesCancel = (index: number): void => {
        series.meta.setState(state => {
            const series = [...state.value]
            const seriesValueBeforeEdit = seriesBeforeEdit[series[index].id]

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

    const handleEditSeriesCommit = (index: number, editedSeries: EditableDataSeries): void => {
        series.meta.setState(state => {
            const series = state.value
            const updatedSeries = { ...editedSeries, valid: true, edit: false }
            const newSeries = replace(series, index, updatedSeries)

            return { ...state, value: newSeries }
        })
    }

    const handleRemoveSeries = (index: number): void => {
        series.meta.setState(state => {
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
