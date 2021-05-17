import classnames from 'classnames'
import React, { useMemo, useState } from 'react'
import { noop } from 'rxjs'

import { Settings } from '@sourcegraph/shared/src/settings/settings'

import { useField } from '../../../../../components/form/hooks/useField'
import { SubmissionErrors, useForm } from '../../../../../components/form/hooks/useForm'
import { useTitleValidator } from '../../../../../components/form/hooks/useTitleValidator'
import { DataSeries } from '../../../../../core/backend/types';
import { InsightTypePrefix } from '../../../../../core/types'
import { useDistinctValue } from '../../../../../hooks/use-distinct-value';
import { CreateInsightFormFields } from '../../types'
import { DEFAULT_ACTIVE_COLOR } from '../form-color-input/FormColorInput';
import { SearchInsightLivePreview } from '../live-preview-chart/SearchInsightLivePreview'
import { SearchInsightCreationForm } from '../search-insight-creation-form/SearchInsightCreationForm'

import styles from './SearchInsightCreationContent.module.scss'
import {
    repositoriesExistValidator,
    repositoriesFieldValidator,
    requiredStepValueField,
    seriesRequired
} from './validators';

const INITIAL_VALUES: CreateInsightFormFields = {
    visibility: 'personal',
    series: [],
    step: 'months',
    stepValue: '2',
    title: '',
    repositories: '',
}

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
    valid: boolean;
}

export interface SearchInsightCreationContentProps {
    /** This component might be used in edit or creation insight case. */
    mode?: 'creation' | 'edit'
    /** Final settings cascade. Used for title field validation. */
    settings?: Settings | null
    /** Initial value for all form fields. */
    initialValue?: CreateInsightFormFields
    /** Custom class name for root form element. */
    className?: string
    /** Submit handler for form element. */
    onSubmit: (values: CreateInsightFormFields) => SubmissionErrors | Promise<SubmissionErrors> | void
    /** Cancel handler. */
    onCancel?: () => void
}

export const SearchInsightCreationContent: React.FunctionComponent<SearchInsightCreationContentProps> = props => {
    const { mode = 'creation', settings, initialValue = INITIAL_VALUES, onSubmit, onCancel = noop, className } = props

    const isEditMode = mode === 'edit'

    const { formAPI, ref, handleSubmit } = useForm<CreateInsightFormFields>({
        initialValues: initialValue,
        onSubmit,
        touched: isEditMode,
    })

    // We can't have two or more insights with the same name, since we rely on name as on id of insights.
    const titleValidator = useTitleValidator({ settings, insightType: InsightTypePrefix.search })

    const title = useField('title', formAPI, { sync: titleValidator })
    const repositories = useField('repositories', formAPI, {
        sync: repositoriesFieldValidator,
        async: repositoriesExistValidator
    })
    const visibility = useField('visibility', formAPI)

    const series = useField('series', formAPI, { sync: seriesRequired })

    const [editSeries, setEditSeries] = useState<(EditDataSeries | undefined)[]>(() => {
        const hasSeries = formAPI.initialValues.series.length;

        if (hasSeries) {
            return formAPI.initialValues.series.map(() => undefined);
        }

        // If we in creation mode we should show first series editor in a first
        // render.
        return [createDefaultEditSeries()]
    })

    const liveSeries = useDistinctValue(
        useMemo<DataSeries[]>(
            () => editSeries
                .map((editSeries, index) => {
                    if (editSeries) {
                        const { valid, ...series } = editSeries
                        return valid ? series : undefined
                    }

                    return series.meta.value[index];
                })
                // eslint-disable-next-line @typescript-eslint/ban-ts-comment
                // @ts-ignore
                .filter<DataSeries>(series => !!series),
            [series, editSeries]
        )
    );

    const step = useField('step', formAPI)
    const stepValue = useField('stepValue', formAPI, { sync: requiredStepValueField })

    // If some fields that needed to run live preview  are invalid
    // we should disabled live chart preview
    const allFieldsForPreviewAreValid =
        repositories.meta.validState === 'VALID' &&
        (series.meta.validState === 'VALID' || liveSeries.length) &&
        stepValue.meta.validState === 'VALID'

    const handleSeriesLiveChange = (liveSeries: DataSeries, valid: boolean, index: number): void => {
        const newEditSeries = [...editSeries];

        newEditSeries[index] = { ...liveSeries, valid }

        setEditSeries(newEditSeries)
    }

    const handleEditSeriesRequest = (index: number): void => {
        const newEditSeries = [...editSeries];

        newEditSeries[index] = series.meta.value[index]
            ? createDefaultEditSeries(series.meta.value[index], true)
            : createDefaultEditSeries()

        setEditSeries(newEditSeries)
    }

    const handleEditSeriesCancel = (index: number): void => {
        const newEditSeries = [...editSeries];

        newEditSeries[index] = undefined
        setEditSeries(newEditSeries)
    }

    const handleEditSeriesCommit = (index: number, editedSeries: DataSeries): void => {
        const newEditedSeries = [...editSeries];
        const newSeries = [
            ...series.input.value.slice(0, index),
            editedSeries,
            ...series.input.value.slice(index + 1),
        ]

        // Remove series from edited cards
        newEditedSeries[index] = undefined

        setEditSeries(newEditedSeries)
        series.input.onChange(newSeries)
    }

    const handleRemoveSeries = (index: number): void => {
        const newSeries = [
            ...series.input.value.slice(0, index),
            ...series.input.value.slice(index + 1),
        ]
        const newEditedSereis = [
            ...editSeries.slice(0, index),
            ...editSeries.slice(index + 1),
        ]

        setEditSeries(newEditedSereis)
        series.input.onChange(newSeries)
    }

    return (
        <div className={classnames(styles.content, className)}>
            <SearchInsightCreationForm
                mode={mode}
                className={styles.contentForm}
                innerRef={ref}
                handleSubmit={handleSubmit}
                submitErrors={formAPI.submitErrors}
                submitting={formAPI.submitting}
                title={title}
                repositories={repositories}
                visibility={visibility}
                series={series}
                step={step}
                stepValue={stepValue}
                onSeriesLiveChange={handleSeriesLiveChange}
                onCancel={onCancel}
                editSeries={editSeries}
                onEditSeriesRequest={handleEditSeriesRequest}
                onEditSeriesCancel={handleEditSeriesCancel}
                onEditSeriesCommit={handleEditSeriesCommit}
                onSeriesRemove={handleRemoveSeries}/>

            <SearchInsightLivePreview
                disabled={!allFieldsForPreviewAreValid}
                repositories={repositories.meta.value}
                series={liveSeries}
                step={step.meta.value}
                stepValue={stepValue.meta.value}
                className={styles.contentLivePreview}
            />
        </div>
    )
}
