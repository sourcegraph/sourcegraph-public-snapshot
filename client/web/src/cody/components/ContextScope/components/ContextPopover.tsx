import React, { useState } from 'react'

import {
    mdiChevronUp,
    mdiCloseCircleOutline,
    mdiClose,
    mdiPlusCircleOutline,
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
    Input,
} from '@sourcegraph/wildcard'

import { TruncatedText } from '../../../../enterprise/insights/components'
import { ContextType, SELECTED } from '../ContextScope'

import { EmptyState } from './EmptyState'
import { repoMockedModel } from './mockedModels'

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
                    {(isEmpty && !isSearching) || isSearchEmpty ? (
                        <EmptyState
                            icon={icon}
                            message={isSearchEmpty ? `No ${itemType} found for '${searchText}'` : emptyMessage}
                        />
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
                                            <Icon aria-hidden={true} svgPath={icon} />{' '}
                                            {
                                                <span
                                                    dangerouslySetInnerHTML={{
                                                        __html: getTintedText(item, searchText),
                                                    }}
                                                />
                                            }
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

// Custom fuzzy search function
export const fuzzySearch = (item: string, search: string): boolean => {
    const escapedSearch = search.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')
    const searchRegex = new RegExp(escapedSearch, 'gi')
    return searchRegex.test(item)
}

export const getTintedText = (item: string, searchText: string) => {
    const searchRegex = new RegExp(`(${searchText})`, 'gi')
    return item.replace(searchRegex, match => `<span class="tinted">${match}</span>`)
}
