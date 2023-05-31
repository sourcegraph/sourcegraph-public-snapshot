import React, { useState } from 'react'

import { mdiFileDocumentOutline, mdiGithub } from '@mdi/js'

import { ChatStatusIndicator } from './components/ChatStatusIndicator'
import { ContextPopover } from './components/ContextPopover'
import { ContextScopePicker } from './components/ContextScopePicker'

import styles from './ContextScope.module.scss'

export const SELECTED = {
    REPOSITORIES: 0,
    NONE: 1,
    AUTOMATIC: 2,
} as const

export type ContextType = typeof SELECTED[keyof typeof SELECTED]

interface ContextScopeProps {}

export const ContextScope: React.FC<ContextScopeProps> = ({}) => {
    const [selectedItem, setSelectedItem] = useState<ContextType>(SELECTED.NONE)

    const handleItemSelected = (itemIndex: ContextType) => {
        setSelectedItem(itemIndex)
    }

    let chatStatus = 'low'
    if (selectedItem === SELECTED.REPOSITORIES) {
        chatStatus = 'medium'
    } else if (selectedItem === SELECTED.NONE) {
        chatStatus = 'high'
    }

    return (
        <div className={styles.wrapper}>
            <ChatStatusIndicator status={chatStatus} />
            {/* <ContextSeparator />
            <div className={styles.title}>Context</div> */}
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
                {selectedItem === SELECTED.AUTOMATIC && 'sourcegraph/sourcegraph —— CodyChat.tsx'}
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
