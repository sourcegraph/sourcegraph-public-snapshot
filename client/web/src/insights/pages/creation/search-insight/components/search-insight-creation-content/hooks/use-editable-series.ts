import { useState } from 'react'

import { useFieldAPI } from '../../../../../../components/form/hooks/useField'
import { DataSeries } from '../../../../../../core/backend/types'
import { useDistinctValue } from '../../../../../../hooks/use-distinct-value'
import { CreateInsightFormFields } from '../../../types'
import { DEFAULT_ACTIVE_COLOR } from '../../form-color-input/FormColorInput'

const createDefaultEditSeries = (series = defaultEditSeries, valid = false): EditDataSeries => ({
    ...series,
    valid,
})

const defaultEditSeries = {
    name: '',
    query: '',
    stroke: DEFAULT_ACTIVE_COLOR,
}

interface EditDataSeries extends DataSeries {
    valid: boolean
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
    editSeries: (CreateInsightFormFields['series'][number] | undefined)[]

    /**
     * Latest valid values of series.
     * */
    liveSeries: DataSeries[]

    /**
     * Handler to listen latest values of particular sereis form.
     * */
    listen: (liveSeries: DataSeries, valid: boolean, index: number) => void

    /**
     * Handlers for CRUD operations over sereis.
     * */
    editRequest: (index: number) => void
    editCommit: (index: number, editedSeries: DataSeries) => void
    cancelEdit: (index: number) => void
    deleteSeries: (index: number) => void
}

/**
 * Implementation of CRUD operation over insight series. Used in form to manage
 * edit, delete, add, and cancel series forms.
 * */
export function useEditableSeries(props: UseEditableSeriesProps): UseEditableSeriesAPI {
    const { series } = props

    const [editSeries, setEditSeries] = useState<(EditDataSeries | undefined)[]>(() => {
        const hasSeries = series.input.value.length

        if (hasSeries) {
            return series.input.value.map(() => undefined)
        }

        // If we in creation mode we should show first series editor in a first
        // render.
        return [createDefaultEditSeries()]
    })

    const liveSeries = useDistinctValue(
        editSeries
            .map((editSeries, index) => {
                if (editSeries) {
                    const { valid, ...series } = editSeries
                    return valid ? series : undefined
                }

                return series.meta.value[index]
            })
            .filter<DataSeries>((series): series is DataSeries => !!series)
    )

    const handleSeriesLiveChange = (liveSeries: DataSeries, valid: boolean, index: number): void => {
        setEditSeries(editSeries => {
            const newEditSeries = [...editSeries]

            newEditSeries[index] = { ...liveSeries, valid }

            return newEditSeries
        })
    }

    const handleEditSeriesRequest = (index: number): void => {
        setEditSeries(editSeries => {
            const newEditSeries = [...editSeries]

            newEditSeries[index] = series.meta.value[index]
                ? createDefaultEditSeries(series.meta.value[index], true)
                : createDefaultEditSeries()

            return newEditSeries
        })
    }

    const handleEditSeriesCancel = (index: number): void => {
        setEditSeries(editSeries => {
            const newEditSeries = [...editSeries]

            newEditSeries[index] = undefined
            setEditSeries(newEditSeries)

            return editSeries
        })
    }

    const handleEditSeriesCommit = (index: number, editedSeries: DataSeries): void => {
        setEditSeries(editSeries => {
            const newEditedSeries = [...editSeries]

            // Remove series from edited cards
            newEditedSeries[index] = undefined

            return newEditedSeries
        })

        const newSeries = [...series.input.value.slice(0, index), editedSeries, ...series.input.value.slice(index + 1)]
        series.input.onChange(newSeries)
    }

    const handleRemoveSeries = (index: number): void => {
        setEditSeries(editSeries => [...editSeries.slice(0, index), ...editSeries.slice(index + 1)])

        const newSeries = [...series.input.value.slice(0, index), ...series.input.value.slice(index + 1)]
        series.input.onChange(newSeries)
    }

    return {
        liveSeries,
        editSeries,
        listen: handleSeriesLiveChange,
        editRequest: handleEditSeriesRequest,
        editCommit: handleEditSeriesCommit,
        cancelEdit: handleEditSeriesCancel,
        deleteSeries: handleRemoveSeries,
    }
}
