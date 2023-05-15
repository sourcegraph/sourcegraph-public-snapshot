import React, { useState } from 'react'

import { mdiFileDocumentOutline, mdiGithub } from '@mdi/js'

import { ContextPopover } from './components/ContextPopover'
import { ContextScopePicker } from './components/ContextScopePicker'

import styles from './ContextScope.module.scss'

export const SELECTED = {
    ORGANIZATIONS: 0,
    REPOSITORIES: 1,
    FILES: 2,
    NONE: 3,
    AUTOMATIC: 4,
} as const

export type ContextType = typeof SELECTED[keyof typeof SELECTED]

interface ContextScopeProps {}

export const ContextScope: React.FC<ContextScopeProps> = ({}) => {
    const [selectedItem, setSelectedItem] = useState<ContextType>(SELECTED.NONE)

    const handleItemSelected = (itemIndex: ContextType) => {
        setSelectedItem(itemIndex)
    }

    return (
        <div className={styles.wrapper}>
            <div className={styles.title}>Context scope</div>
            <ContextSeparator />
            <ContextScopePicker onSelect={handleItemSelected} selected={selectedItem} />
            <ContextSeparator />
            <div
                className="d-flex"
                style={{
                    flexGrow: 1,
                }}
            >
                {selectedItem === SELECTED.FILES && <ItemFiles />}
                {selectedItem === SELECTED.REPOSITORIES && <ItemRepos />}
            </div>
        </div>
    )
}

const ContextSeparator = () => <div className={styles.separator} />

const ItemRepos: React.FC = () => {
    const mockedRepoNames: string[] = []

    return (
        <ContextPopover
            header="Add repositories to the scope"
            icon={mdiGithub}
            emptyMessage="Start by adding repositories to the scope."
            inputPlaceholder="Search for a repository..."
            items={mockedRepoNames}
            contextType={SELECTED.REPOSITORIES}
            itemType="repositories"
        />
    )
}

const ItemFiles: React.FC = () => {
    const mockedFileNames: string[] = []

    return (
        <ContextPopover
            header="Add files to the scope"
            icon={mdiFileDocumentOutline}
            emptyMessage="Start by adding files to the scope."
            inputPlaceholder="Search for a file..."
            items={mockedFileNames}
            contextType={SELECTED.FILES}
            itemType="files"
        />
    )
}
