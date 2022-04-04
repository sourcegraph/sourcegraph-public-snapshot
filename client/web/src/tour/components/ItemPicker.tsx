import React from 'react'

import classNames from 'classnames'
import CloseIcon from 'mdi-react/CloseIcon'

import { Button } from '@sourcegraph/wildcard'

import styles from './ItemPicker.module.scss'

interface ItemPickerProps<TItem> {
    items: TItem[]
    onClose: () => void
    onSelect: (language: TItem) => void
}

/**
 * ItemPicker component. Displays a closable block with list of items passed.
 */
export const ItemPicker = <TItem extends string>({
    items,
    onClose,
    onSelect,
}: ItemPickerProps<TItem>): React.ReactElement => (
    <div>
        <div className="d-flex justify-content-between">
            <p className="mt-2">Please select a language:</p>
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
