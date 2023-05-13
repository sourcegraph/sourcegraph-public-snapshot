import React, { useState } from 'react'

import { mdiChevronUp } from '@mdi/js'

import { Icon, Popover, PopoverTrigger, PopoverContent, Position, Button } from '@sourcegraph/wildcard'

import { ContextScopePicker } from './components/ContextScopePicker'

import styles from './ContextScope.module.scss'

export const SELECTED = {
    ORGANIZATIONS: 0,
    REPOSITORIES: 1,
    FILES: 2,
    NONE: 3,
} as const

export type SelectedType = typeof SELECTED[keyof typeof SELECTED]

interface ContextScopeProps {}

export const ContextScope: React.FC<ContextScopeProps> = ({}) => {
    const [selectedItem, setSelectedItem] = useState<SelectedType>(SELECTED.NONE)

    const handleItemSelected = (itemIndex: SelectedType) => {
        setSelectedItem(itemIndex)
    }

    return (
        <div className={styles.wrapper}>
            <div className={styles.title}>Context scope</div>
            <ContextSeparator />
            <ContextScopePicker onSelect={handleItemSelected} selected={selectedItem} />
            <ContextSeparator />
            <div
                style={{
                    display: 'flex',
                    flexGrow: 1,
                }}
            >
                {selectedItem === SELECTED.FILES && <ItemFiles />}
                {selectedItem === SELECTED.REPOSITORIES && <div>react, ant-design, sourcegraph, and 3 more</div>}
            </div>
        </div>
    )
}

const ContextSeparator = () => <div className={styles.separator} />

const ItemFiles: React.FC = () => {
    return (
        <Popover>
            <PopoverTrigger
                as={Button}
                outline={false}
                style={{
                    display: 'flex',
                    flexGrow: 1,
                    justifyContent: 'space-between',
                    padding: 0,
                    alignItems: 'center',
                }}
            >
                <div>CodyChat.tsx, CodyChatView.tsx, Cody.ts, 3 more</div>
                <Icon aria-hidden={true} svgPath={mdiChevronUp} />
            </PopoverTrigger>
            <PopoverContent position={Position.topStart}>Hi</PopoverContent>
        </Popover>
    )
}

export default ContextScope
