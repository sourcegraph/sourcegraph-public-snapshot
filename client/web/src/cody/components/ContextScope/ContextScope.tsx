import React, { useState } from 'react'

import {
    mdiChevronUp,
    mdiCloseCircleOutline,
    mdiFileDocumentOutline,
    mdiFileOutline,
    mdiClose,
    mdiGit,
    mdiMinusCircleOutline,
    mdiGithub,
} from '@mdi/js'
import classNames from 'classnames'

import {
    Icon,
    Popover,
    PopoverTrigger,
    PopoverContent,
    Position,
    Button,
    Card,
    Text,
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
    contextType: ContextType
}> = ({ header, icon, emptyMessage, inputPlaceholder, items, contextType }) => {
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
                className="d-flex justify-content-between p-0 align-items-center w-100"
            >
                <div>CodyChat.tsx, CodyChatView.tsx, Cody.ts, 3 more</div>
                <Icon aria-hidden={true} svgPath={mdiChevronUp} />
            </PopoverTrigger>

            <PopoverContent position={Position.topStart}>
                <Card>
                    <div className={classNames('justify-content-between', 'header')}>
                        {header}
                        <Button onClick={() => setIsPopoverOpen(false)} variant="icon" aria-label="Close">
                            <Icon aria-hidden={true} svgPath={mdiClose} />
                        </Button>
                    </div>
                    {isEmpty ? (
                        <EmptyState icon={icon} message={emptyMessage} />
                    ) : (
                        <>
                            <div className="itemsContainer">
                                {currentItems?.map((item, index) => (
                                    <div
                                        key={index}
                                        className={classNames('d-flex justify-content-between flex-row p-1', 'item')}
                                    >
                                        <div>
                                            <Icon aria-hidden={true} svgPath={icon} /> {item}
                                        </div>
                                        <div className="d-flex align-items-center itemRight">
                                            {contextType !== SELECTED.FILES && (
                                                <>
                                                    <Icon aria-hidden={true} svgPath={mdiGithub} />{' '}
                                                    <Text size="small" className="m-0">
                                                        sourcegraph/sourcegraph/client/cody-ui/src
                                                    </Text>
                                                </>
                                            )}

                                            <Button
                                                className="pl-1"
                                                variant="icon"
                                                onClick={() => handleRemoveItem(index)}
                                            >
                                                <Icon aria-hidden={true} svgPath={mdiMinusCircleOutline} />
                                            </Button>
                                        </div>
                                    </div>
                                ))}
                            </div>

                            <Button
                                className={classNames('d-flex justify-content-between', 'itemClear')}
                                variant="icon"
                                onClick={handleClearAll}
                            >
                                Clear all from the scope.
                                <Icon aria-hidden={true} svgPath={mdiCloseCircleOutline} />
                            </Button>
                        </>
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
        'almeidapaulooliveira/ant-design',
        'sourcegraph/sourcegraph',
        'almeidapaulooliveira/react-custom-scrollbars',
        'almeidapaulooliveira/react-packages',
        'bartonhammond/meteor-slingshot',
    ]

    return (
        <PopoverComponent
            header="Add repositories to the scope"
            icon={mdiGithub}
            emptyMessage="Start by adding repositories to the scope."
            inputPlaceholder="Search for a repository..."
            items={mockedRepoNames}
            contextType={SELECTED.FILES}
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
            contextType={SELECTED.REPOSITORIES}
        />
    )
}
