import React, { useState } from 'react'

import {
    mdiChevronUp,
    mdiCloseCircleOutline,
    mdiFileDocumentOutline,
    mdiFileOutline,
    mdiClose,
    mdiPlusCircleOutline,
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

import { TruncatedText } from '../../../enterprise/insights/components'

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

// Custom fuzzy search function
const fuzzySearch = (item, search) => {
    const itemLowerCase = item.toLowerCase()
    const searchLowerCase = search.toLowerCase()

    let lastIndex = -1
    for (let i = 0; i < searchLowerCase.length; i++) {
        const index = itemLowerCase.indexOf(searchLowerCase[i], lastIndex + 1)
        if (index === -1) {
            return false
        }
        lastIndex = index
    }
    return true
}

const ContextSeparator = () => <div className={styles.separator} />

const PopoverComponent: React.FC<{
    header: string
    icon: string
    emptyMessage: string
    inputPlaceholder: string
    items?: string[]
    contextType: ContextType
    itemType: string
}> = ({ header, icon, emptyMessage, inputPlaceholder, items, contextType, itemType }) => {
    const [isPopoverOpen, setIsPopoverOpen] = useState(false)
    const [currentItems, setCurrentItems] = useState<string[] | undefined>(items)
    const [searchText, setSearchText] = useState('') // Add state for the search text

    const clearSearchText = () => {
        setSearchText('')
    }

    const handleSearch = (event: React.ChangeEvent<HTMLInputElement>) => {
        setSearchText(event.target.value)
    }

    const handleClearAll = () => {
        setCurrentItems([])
    }

    const handleAddAll = () => {
        if (filteredItems && filteredItems.length > 0) {
            setCurrentItems(prevItems => (prevItems ? [...prevItems, ...filteredItems] : filteredItems))
            clearSearchText() // Clear the search text after adding all items
        }
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

    const handleAddItem = (index: number) => {
        if (filteredItems) {
            const selectedItem = filteredItems[index]
            setCurrentItems(prevItems => (prevItems ? [...prevItems, selectedItem] : [selectedItem]))
        }
    }

    const filteredItems = repoMockedModel
        .filter(item => !currentItems?.includes(item)) // Exclude items already present in currentItems
        .filter(item => fuzzySearch(item, searchText))

    const isEmpty = !currentItems || currentItems.length === 0
    const isSearching = searchText.length > 0

    return (
        <Popover isOpen={isPopoverOpen} onOpenChange={event => setIsPopoverOpen(event.isOpen)}>
            <PopoverTrigger
                as={Button}
                outline={false}
                className="d-flex justify-content-between p-0 align-items-center w-100 trigger"
            >
                <div className={classNames(isEmpty && 'emptyTrigger', 'innerTrigger')}>
                    {isEmpty ? (
                        `${header}...`
                    ) : (
                        <TruncatedText>
                            {currentItems.length} {itemType} ({currentItems?.join(', ')})
                        </TruncatedText>
                    )}
                </div>
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
                    {isEmpty && !isSearching ? (
                        <EmptyState icon={icon} message={emptyMessage} />
                    ) : (
                        <>
                            <div className="itemsContainer">
                                {(isSearching ? filteredItems : currentItems)?.map((item, index) => (
                                    <div
                                        key={index}
                                        className={classNames(
                                            'd-flex justify-content-between flex-row p-1 rounded-lg',
                                            'item'
                                        )}
                                    >
                                        <div>
                                            <Icon aria-hidden={true} svgPath={icon} /> {item}
                                        </div>
                                        <div className="d-flex align-items-center itemRight">
                                            {contextType === SELECTED.FILES && (
                                                <>
                                                    <Icon aria-hidden={true} svgPath={mdiGithub} />{' '}
                                                    <Text size="small" className="m-0">
                                                        sourcegraph/sourcegraph/client/cody-ui/src
                                                    </Text>
                                                </>
                                            )}

                                            {isSearching ? (
                                                <Button
                                                    className="pl-1"
                                                    variant="icon"
                                                    onClick={() => handleAddItem(index)}
                                                >
                                                    <Icon aria-hidden={true} svgPath={mdiPlusCircleOutline} />
                                                </Button>
                                            ) : (
                                                <Button
                                                    className="pl-1"
                                                    variant="icon"
                                                    onClick={() => handleRemoveItem(index)}
                                                >
                                                    <Icon aria-hidden={true} svgPath={mdiMinusCircleOutline} />
                                                </Button>
                                            )}
                                        </div>
                                    </div>
                                ))}
                            </div>

                            {isSearching ? (
                                <Button
                                    className={classNames('d-flex justify-content-between', 'itemClear')}
                                    variant="icon"
                                    onClick={handleAddAll}
                                >
                                    Add all to the scope.
                                    <Icon aria-hidden={true} svgPath={mdiPlusCircleOutline} />
                                </Button>
                            ) : (
                                <Button
                                    className={classNames('d-flex justify-content-between', 'itemClear')}
                                    variant="icon"
                                    onClick={handleClearAll}
                                >
                                    Clear all from the scope.
                                    <Icon aria-hidden={true} svgPath={mdiCloseCircleOutline} />
                                </Button>
                            )}
                        </>
                    )}

                    <div className="footer">
                        <Input
                            role="combobox"
                            autoFocus={true}
                            autoComplete="off"
                            spellCheck="false"
                            placeholder={inputPlaceholder}
                            variant="small"
                            value={searchText}
                            onChange={handleSearch}
                        />
                    </div>
                </Card>
            </PopoverContent>
        </Popover>
    )
}

const repoMockedModel = [
    'almeidapaulooliveira/ant-design',
    'sourcegraph/sourcegraph',
    'almeidapaulooliveira/react-custom-scrollbars',
    'almeidapaulooliveira/react-packages',
    'bartonhammond/meteor-slingshot',
    'primer/react',
    'redis/redis',
    'facebook/react',
    'npm/npm',
    'docker/app',
    'kubernetes/dashboard',
    'kubernetes/website',
    'kubernetes/kubernetes',
    'nginx/docker-nginx',
    'opencontainers/runc',
    'opencontainers/image-spec',
    'ant-design/ant-design',
    'ant-design/ant-design-pro',
    'ant-design/ant-design-pro-layout',
    'ant-design/ant-design-icons',
    'ant-design/ant-design-pro-blocks',
    'ant-design/ant-design-pro-cli',
    'ant-design/ant-design-pro-site',
    'ant-design/ant-design-pro-table',
    'ant-design/ant-design-pro-form',
    'ant-design/ant-design-pro-list',
    'ant-design/ant-design-pro-field',
    'ant-design/ant-design-pro-descriptions',
]

const ItemRepos: React.FC = () => {
    const mockedRepoNames = repoMockedModel.slice(0, 6)

    return (
        <PopoverComponent
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
            contextType={SELECTED.FILES}
            itemType="files"
        />
    )
}
