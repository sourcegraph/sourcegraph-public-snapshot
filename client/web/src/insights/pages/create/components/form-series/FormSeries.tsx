import classnames from 'classnames';
import React, { ReactElement, useCallback, useState } from 'react';

import { DataSeries } from '../../types';
import { FormSeriesInput } from '../form-series-input/FormSeriesInput';

import styles from './FormSeries.module.scss'

export interface FormSeriesProps {
    series?: DataSeries[];
    onChange: (series: DataSeries[]) => void;
}

export function FormSeries(props: FormSeriesProps): ReactElement {
    const { series = [], onChange } = props;

    const [editSeriesIndexes, setEditSeriesIndex] = useState<number[]>([]);
    const [newSeriesEdit, setNewSeriesEdit] = useState(false);

    const handleAddClick = useCallback(() => {
        setNewSeriesEdit(true);
    }, [setNewSeriesEdit])

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

    const handleEditSeries = (index: number, editedSeries:DataSeries): void => {
        const newSeries = [...series];

        newSeries[index] = editedSeries;
        setEditSeriesIndex(indexes => indexes.filter(currentIndex => currentIndex !== index))
        onChange(newSeries);
    }

    const handleEditSeriesForm = (index: number): void => {
        setEditSeriesIndex([...editSeriesIndexes, index])
    }

    if (series.length === 0) {
        return (
            <FormSeriesInput
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
                            /* eslint-disable-next-line react/jsx-no-bind */
                            onSubmit={series => handleEditSeries(index, series)}
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
                    onSubmit={handleSubmitNewSeries}
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
        <section
            // eslint-disable-next-line jsx-a11y/no-noninteractive-tabindex
            tabIndex={0}
            onPointerUp={onEdit}
            className={classnames(styles.formSeriesCard, className)}>

            <div className={styles.formSeriesCardContent}>

                <h4 className={styles.formSeriesCardName}>{name}</h4>
                <p className={classnames(styles.formSeriesCardQuery, 'text-muted')}>{query}</p>
            </div>

            {/* eslint-disable-next-line react/forbid-dom-props */}
            <div style={{ color }} className={styles.formSeriesCardColor}/>
        </section>
    )
}
