import React, { useState, useEffect } from 'react'

import {
    mdiChevronUp,
    mdiCloseCircleOutline,
    mdiClose,
    mdiPlusCircleOutline,
    mdiMinusCircleOutline,
    mdiGithub,
    mdiCloseCircle,
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
    Input,
} from '@sourcegraph/wildcard'

import { TruncatedText } from '../../../../enterprise/insights/components'
import { ContextType, SELECTED } from '../ContextScope'

import { repoMockedModel, filesMockedModel } from './mockedModels'

import styles from './ContextComponents.module.scss'

export const ContextPopover: React.FC<{
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
    const [searchText, setSearchText] = useState('')

    useEffect(() => {
        setCurrentItems(items)
        clearSearchText()
    }, [items])

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

    const filteredItems =
        contextType === SELECTED.REPOSITORIES
            ? repoMockedModel.filter(item => !currentItems?.includes(item) && fuzzySearch(item, searchText))
            : filesMockedModel.filter(item => !currentItems?.includes(item) && fuzzySearch(item, searchText))

    const isSearching = searchText.length > 0
    const isSearchEmpty = isSearching && filteredItems.length === 0
    const isEmpty = !currentItems || currentItems.length === 0

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
                            {currentItems.length} {itemType} ({currentItems?.map(item => getFileName(item)).join(', ')})
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
                    {(isEmpty && !isSearching) || isSearchEmpty ? (
                        <EmptyState
                            icon={icon}
                            message={isSearchEmpty ? `No ${itemType} found for '${searchText}'` : emptyMessage}
                        />
                    ) : (
                        <>
                            <div className="itemsContainer">
                                {(isSearching ? filteredItems : currentItems)?.map((item, index) => (
                                    <ContextItem
                                        item={item}
                                        icon={icon}
                                        searchText={searchText}
                                        contextType={contextType}
                                        handleAddItem={() => handleAddItem(index)}
                                        handleRemoveItem={() => handleRemoveItem(index)}
                                    />
                                ))}
                            </div>

                            <ContextActions
                                isSearching={isSearching}
                                handleAddAll={handleAddAll}
                                handleClearAll={handleClearAll}
                            />
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
                        {isSearching && (
                            <Button className="clearButton" variant="icon" onClick={clearSearchText} aria-label="Clear">
                                <Icon aria-hidden={true} svgPath={mdiCloseCircle} />
                            </Button>
                        )}
                    </div>
                </Card>
            </PopoverContent>
        </Popover>
    )
}

const ContextItem: React.FC<{
    item: string
    icon: string
    searchText: string
    contextType: ContextType
    handleAddItem: () => void
    handleRemoveItem: () => void
}> = ({ item, icon, searchText, contextType, handleAddItem, handleRemoveItem }) => (
    <div className={classNames('d-flex justify-content-between flex-row p-1 rounded-lg', 'item')}>
        <div>
            <Icon aria-hidden={true} svgPath={icon} />{' '}
            <span
                dangerouslySetInnerHTML={{
                    __html: getTintedText(contextType === SELECTED.FILES ? getFileName(item) : item, searchText),
                }}
            />
        </div>
        <div className="d-flex align-items-center itemRight">
            {contextType === SELECTED.FILES && (
                <>
                    <Icon aria-hidden={true} svgPath={mdiGithub} />{' '}
                    <Text size="small" className="m-0">
                        <span dangerouslySetInnerHTML={{ __html: getTintedText(getPath(item), searchText) }} />
                    </Text>
                </>
            )}
            <ItemAction
                isSearching={searchText.length > 0}
                handleAddItem={handleAddItem}
                handleRemoveItem={handleRemoveItem}
            />
        </div>
    </div>
)

const ItemAction: React.FC<{
    isSearching: boolean
    handleAddItem: () => void
    handleRemoveItem: () => void
}> = ({ isSearching, handleAddItem, handleRemoveItem }) => (
    <Button className="pl-1" variant="icon" onClick={isSearching ? handleAddItem : handleRemoveItem}>
        <Icon aria-hidden={true} svgPath={isSearching ? mdiPlusCircleOutline : mdiMinusCircleOutline} />
    </Button>
)

const ContextActions: React.FC<{
    isSearching: boolean
    handleAddAll: () => void
    handleClearAll: () => void
}> = ({ isSearching, handleAddAll, handleClearAll }) => {
    const buttonLabel = isSearching ? 'Add all to the scope.' : 'Clear all from the scope.'
    const buttonIcon = isSearching ? mdiPlusCircleOutline : mdiCloseCircleOutline

    return (
        <Button
            className={classNames('d-flex justify-content-between', 'itemClear')}
            variant="icon"
            onClick={isSearching ? handleAddAll : handleClearAll}
        >
            {buttonLabel}
            <Icon aria-hidden={true} svgPath={buttonIcon} />
        </Button>
    )
}

/**
 * Displays an empty state icon and message.
 */
const EmptyState: React.FC<{ icon: string; message: string }> = ({ icon, message }) => (
    <div className={classNames('d-flex align-items-center justify-content-center flex-column', styles.emptyState)}>
        <svg height={40} width={40} viewBox="0 0 24 24">
            <path d={icon} fill="currentColor" />
        </svg>

        <Text size="small" className="m-0 d-flex text-center">
            {message}
        </Text>
    </div>
)

/**
 * Helper fuctions for search and filtering.
 */
export const fuzzySearch = (item: string, search: string): boolean => {
    const escapedSearch = search.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')
    const searchRegex = new RegExp(escapedSearch, 'gi')
    return searchRegex.test(item)
}

export const getTintedText = (item: string, searchText: string) => {
    const searchRegex = new RegExp(`(${searchText})`, 'gi')
    return item.replace(searchRegex, match => `<span class="tinted">${match}</span>`)
}

export const getFileName = (path: string) => {
    const parts = path.split('/')
    return parts[parts.length - 1]
}

export const getPath = (path: string) => {
    const parts = path.split('/')
    return parts.slice(0, parts.length - 1).join('/')
}
