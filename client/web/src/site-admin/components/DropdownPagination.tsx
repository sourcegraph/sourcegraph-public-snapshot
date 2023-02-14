import React, { useState, useCallback } from 'react'

import { mdiChevronRight, mdiChevronLeft } from '@mdi/js'
import classNames from 'classnames'

import { Button, Text, Icon, Popover, PopoverContent, PopoverOpenEvent, PopoverTrigger } from '@sourcegraph/wildcard'

export interface DropdownPaginationProps {
    total: number
    offset: number
    limit: number
    options?: ({ limit: number; label: string } | number)[]
    onOffsetChange: (offset: number) => void
    onLimitChange: (limit: number) => void
    formatLabel?: (start: number, end: number, total: number) => string
    className?: string
}

export const DropdownPagination: React.FunctionComponent<DropdownPaginationProps> = ({
    limit,
    offset,
    total,
    options,
    onLimitChange,
    onOffsetChange,
    formatLabel = (start: number, end: number, total: number) => `Showing ${start}-${end} items of ${total}`,
    className,
}) => {
    const [isOpen, setIsOpen] = useState<boolean>(false)
    const handleOpenChange = useCallback((event: PopoverOpenEvent): void => {
        setIsOpen(event.isOpen)
    }, [])

    const handlePreviousPage = useCallback(
        () => onOffsetChange(Math.max(offset - limit, 0)),
        [limit, offset, onOffsetChange]
    )
    const handleNextPage = useCallback(() => {
        const newOffset = offset + limit
        if (total > newOffset) {
            onOffsetChange(newOffset)
        }
    }, [limit, offset, total, onOffsetChange])

    return (
        <div className={classNames('d-flex justify-content-between align-items-center', className)}>
            <Button className="mr-2" onClick={handlePreviousPage}>
                <Icon aria-label="Show previous icon" svgPath={mdiChevronLeft} />
            </Button>
            <Popover isOpen={isOpen} onOpenChange={handleOpenChange}>
                <PopoverTrigger as={Text} className="m-0 p-0 cursor-pointer">
                    <Text as="span">{formatLabel(Math.min(offset + 1), Math.min(offset + limit, total), total)}</Text>
                </PopoverTrigger>
                <PopoverContent focusLocked={false}>
                    <ul className="list-unstyled mb-0">
                        {options
                            ?.map(option =>
                                typeof option === 'number'
                                    ? { limit: option, label: `Show ${option} per page` }
                                    : option
                            )
                            .map(item => (
                                <Button
                                    className="d-flex cursor-pointer"
                                    key={item.limit}
                                    variant="link"
                                    as="li"
                                    outline={true}
                                    onClick={() => {
                                        onLimitChange(item.limit)
                                        setIsOpen(false)
                                    }}
                                >
                                    <span
                                        className={classNames(
                                            item.limit === limit ? 'font-weight-medium' : 'font-weight-normal ml-3'
                                        )}
                                    >
                                        {item.limit === limit && '✓'} {item.label}
                                    </span>
                                </Button>
                            ))}
                    </ul>
                </PopoverContent>
            </Popover>
            <Button onClick={handleNextPage}>
                <Icon aria-label="Show next icon" svgPath={mdiChevronRight} />
            </Button>
        </div>
    )
}
