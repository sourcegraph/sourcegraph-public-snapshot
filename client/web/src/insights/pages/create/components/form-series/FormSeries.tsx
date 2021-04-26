import classnames from 'classnames';
import React, { forwardRef, ReactElement, useCallback, useImperativeHandle, useRef, useState } from 'react';

import { DataSeries } from '../../types';
import { FormSeriesInput, FormSeriesInputAPI } from '../form-series-input/FormSeriesInput';

import styles from './FormSeries.module.scss'

export interface FormSeriesProps {
    name: string;
    series?: DataSeries[];
    onChange: (series: DataSeries[]) => void;
}

export interface FormSeriesReferenceAPI {
    name: string;
    focus: () => void;
}

export const FormSeries = forwardRef<FormSeriesReferenceAPI, FormSeriesProps>(
    (props, reference) => {
        const { name, series = [], onChange } = props;

        const [editSeriesIndexes, setEditSeriesIndex] = useState<number[]>([]);
        const [newSeriesEdit, setNewSeriesEdit] = useState(false);

        const seriesInputReference = useRef<FormSeriesInputAPI>(null);

        const handleAddClick = useCallback(() => {
            setNewSeriesEdit(true);
        }, [setNewSeriesEdit])

        const handleCancelNewSeries = useCallback(
            () => {
                setNewSeriesEdit(false)
            },
            [setNewSeriesEdit]
        )

        const handleSubmitNewSeries = useCallback(
            (newSeries: DataSeries) => {
                // Close series input in case if we add another series
                if (newSeriesEdit) {
                    setNewSeriesEdit(false)
                }

                onChange([...series, newSeries])
            },
            [series, newSeriesEdit, setNewSeriesEdit, onChange]
        );

        const handleSubmitSeries = (index: number, editedSeries:DataSeries): void => {
            const newSeries = [...series];

            newSeries[index] = editedSeries;
            setEditSeriesIndex(indexes => indexes.filter(currentIndex => currentIndex !== index))
            onChange(newSeries);
        }

        const handleEditSeriesForm = (index: number): void => {
            setEditSeriesIndex([...editSeriesIndexes, index])
        }

        const handleCancelEditSeries = (index: number): void => {
            setEditSeriesIndex(indexes => indexes.filter(currentIndex => currentIndex !== index))
        }

        useImperativeHandle(reference, () => ({
            name,
            focus: () => {
                seriesInputReference.current?.focus()
            }
        }))

        if (series.length === 0) {
            return (
                <FormSeriesInput
                    innerRef={seriesInputReference}
                    className={styles.formSeriesInput}
                    onSubmit={handleSubmitNewSeries}/>
            )
        }

        return (
            <div className={classnames(styles.formSeries)}>
                {
                    series.map((line, index) =>
                        editSeriesIndexes.includes(index)
                            ? <FormSeriesInput
                                key={`${line.name}-${index}`}
                                cancel={true}
                                /* eslint-disable-next-line react/jsx-no-bind */
                                onSubmit={series => handleSubmitSeries(index, series)}
                                /* eslint-disable-next-line react/jsx-no-bind */
                                onCancel={() => handleCancelEditSeries(index)}
                                className={classnames(styles.formSeriesInput, styles.formSeriesItem)}
                                {...line}/>
                            : <SeriesCard
                                key={`${line.name}-${index}`}
                                /* eslint-disable-next-line react/jsx-no-bind */
                                onEdit={() => handleEditSeriesForm(index)}
                                className={styles.formSeriesItem}
                                {...line}/>
                    )
                }

                { newSeriesEdit &&
                <FormSeriesInput
                    innerRef={seriesInputReference}
                    cancel={true}
                    onSubmit={handleSubmitNewSeries}
                    onCancel={handleCancelNewSeries}
                    className={classnames(styles.formSeriesInput, styles.formSeriesItem)}/>
                }

                { !newSeriesEdit &&
                <button
                    type='button'
                    onClick={handleAddClick}
                    className={classnames(styles.formSeriesItem, styles.formSeriesAddButton ,'button')}>

                    + Add another data series
                </button>
                }
            </div>
        )
    }
)

interface SeriesCardProps {
    name: string;
    query: string;
    color: string;
    className?: string;
    onEdit?: () => void;
    onRemove?: () => void;
}

function SeriesCard (props: SeriesCardProps): ReactElement {
    const { name, query, color, className, onEdit } = props;

    return (
        <button
            type='button'
            // eslint-disable-next-line jsx-a11y/no-noninteractive-tabindex
            tabIndex={0}
            onClick={onEdit}
            className={classnames(styles.formSeriesCard, className, 'btn')}>

            <div className={styles.formSeriesCardContent}>

                <h4 className={styles.formSeriesCardName}>{name}</h4>
                <p className={classnames(styles.formSeriesCardQuery, 'text-muted')}>{query}</p>
            </div>

            {/* eslint-disable-next-line react/forbid-dom-props */}
            <div style={{ color }} className={styles.formSeriesCardColor}/>
        </button>
    )
}
