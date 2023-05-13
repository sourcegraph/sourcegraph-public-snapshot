import React, { useState } from 'react'

import {
    mdiChevronUp,
    mdiCloseCircleOutline,
    mdiFileDocumentOutline,
    mdiFileOutline,
    mdiClose,
    mdiGit,
    mdiMinusCircleOutline,
} from '@mdi/js'

import {
    Icon,
    Popover,
    PopoverTrigger,
    PopoverContent,
    Position,
    Button,
    Card,
    CardBody,
    CardHeader,
    Input,
    Label,
} from '@sourcegraph/wildcard'

import { ContextScopePicker } from './components/ContextScopePicker'
import { EmptyState } from './components/EmptyState'

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
                {selectedItem === SELECTED.REPOSITORIES && <ItemRepos />}
            </div>
        </div>
    )
}

const ContextSeparator = () => <div className={styles.separator} />

const PopoverComponent: React.FC<{
    header: string
    icon: string
    emptyMessage: string
    inputPlaceholder: string
    items?: string[]
}> = ({ header, icon, emptyMessage, inputPlaceholder, items }) => {
    const [isPopoverOpen, setIsPopoverOpen] = useState(false)
    const [currentItems, setCurrentItems] = useState<string[] | undefined>(items)

    const handleClearAll = () => {
        setCurrentItems([])
    }

    const handleRemoveItem = (index: number) => {
        setCurrentItems(prevItems => {
            if (prevItems) {
                const updatedItems = [...prevItems]
                updatedItems.splice(index, 1)
                return updatedItems
            }
            return prevItems
        })
    }

    const isEmpty = !currentItems || currentItems.length === 0

    return (
        <Popover isOpen={isPopoverOpen} onOpenChange={event => setIsPopoverOpen(event.isOpen)}>
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

            <PopoverContent position={Position.topStart}>
                <Card>
                    <div className="header">
                        {header}
                        <Button onClick={() => setIsPopoverOpen(false)} variant="icon" aria-label="Close">
                            <Icon aria-hidden={true} svgPath={mdiClose} />
                        </Button>
                    </div>
                    <div>
                        {isEmpty ? (
                            <EmptyState icon={icon} message={emptyMessage} />
                        ) : (
                            <div className="itemsContainer">
                                {currentItems?.map((item, index) => (
                                    <div key={index} className="item">
                                        <div>
                                            <Icon aria-hidden={true} svgPath={icon} /> {item}
                                        </div>
                                        <Button variant="icon" onClick={() => handleRemoveItem(index)}>
                                            <Icon aria-hidden={true} svgPath={mdiMinusCircleOutline} />
                                        </Button>
                                    </div>
                                ))}
                            </div>
                        )}
                    </div>
                    {!isEmpty && (
                        <div className="itemClear">
                            <Button variant="icon" size="sm" onClick={handleClearAll}>
                                Clear all from the scope.
                            </Button>
                            <Button variant="icon" onClick={handleClearAll}>
                                <Icon aria-hidden={true} svgPath={mdiCloseCircleOutline} />
                            </Button>
                        </div>
                    )}

                    <div className="footer">
                        <Input placeholder={inputPlaceholder} variant="small" />
                    </div>
                </Card>
            </PopoverContent>
        </Popover>
    )
}

const ItemRepos: React.FC = () => {
    const mockedRepoNames = [
        'ant-design',
        'sourcegraph',
        'react-custom-scrollbars',
        'react-packages',
        'meteor-slingshot',
    ]

    return (
        <PopoverComponent
            header="Add repositories to the scope"
            icon={mdiGit}
            emptyMessage="Start by adding repositories to the scope."
            inputPlaceholder="Search for a repository..."
            items={mockedRepoNames}
        />
    )
}

const ItemFiles: React.FC = () => {
    const mockedFileNames = [
        'CodyChat.tsx',
        'CodyChatView.tsx',
        'Cody.ts',
        'CodyLogo.tsx',
        'CodyModal.ts',
        'CustomScroller.tsx',
    ]

    return (
        <PopoverComponent
            header="Add files to the scope"
            icon={mdiFileDocumentOutline}
            emptyMessage="Start by adding files to the scope."
            inputPlaceholder="Search for a file..."
            items={mockedFileNames}
        />
    )
}
