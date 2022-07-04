import React from 'react'

import classNames from 'classnames'

import { Button, Text, Icon } from '@sourcegraph/wildcard'

import styles from './ItemPicker.module.scss'
import { mdiClose } from "@mdi/js";

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
            <Text className="mt-0 mb-1">{title}</Text>
            <Icon onClick={onClose} svgPath={mdiClose} inline={false} aria-hidden={true} height="1rem" width="1rem" />
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
