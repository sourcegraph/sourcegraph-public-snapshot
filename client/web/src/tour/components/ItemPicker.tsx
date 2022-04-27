import React from 'react'

import classNames from 'classnames'
import CloseIcon from 'mdi-react/CloseIcon'

import { Button } from '@sourcegraph/wildcard'

import styles from './ItemPicker.module.scss'

interface ItemPickerProps<TItem> {
    title: string
    items: TItem[]
    onClose: () => void
    onSelect: (language: TItem) => void
    className?: string
}

/**
 * ItemPicker component. Displays a closable block with list of items passed.
 */
export const ItemPicker = <TItem extends string>({
    title,
    items,
    onClose,
    onSelect,
    className,
}: ItemPickerProps<TItem>): React.ReactElement => (
    <div className={className}>
        <div className="d-flex justify-content-between">
            <p className="mt-0 mb-1">{title}</p>
            <CloseIcon onClick={onClose} size="1rem" />
        </div>
        <div className="d-flex flex-wrap">
            {items.map(language => (
                <Button
                    key={language}
                    className={classNames('mr-1 my-1', styles.item)}
                    onClick={() => onSelect(language)}
                    size="sm"
                >
                    {language}
                </Button>
            ))}
        </div>
    </div>
)
