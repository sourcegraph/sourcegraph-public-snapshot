import React, { useState, useMemo, useCallback, useEffect } from 'react'

import {
    mdiMenuUp,
    mdiMenuDown,
    mdiArrowRightTop,
    mdiArrowRightBottom,
    mdiChevronDown,
    mdiChevronRight,
    mdiChevronLeft,
    mdiFilterRemoveOutline,
} from '@mdi/js'
import classNames from 'classnames'
import { isEqual } from 'lodash'

import {
    Icon,
    Text,
    Checkbox,
    PopoverTrigger,
    PopoverContent,
    Popover,
    Position,
    Button,
    Tooltip,
    type PopoverOpenEvent,
    Input,
    Select,
} from '@sourcegraph/wildcard'

import { DateRangeSelect, type DateRangeSelectProps } from './DateRangeSelect'

import styles from './Table.module.scss'

interface TextFilterProps {
    type: 'text'
    placeholder: string
    value?: string
    onChange?: (value: string) => void
}

interface SelectFilterProps {
    type: 'select'
    options: {
        label: string
        value: string
    }[]
    value?: string
    onChange?: (value: string) => void
}

type ColumnFilterProps = TextFilterProps | SelectFilterProps | (DateRangeSelectProps & { type: 'date-range' })

const ColumnFilter: React.FunctionComponent<ColumnFilterProps> = props => {
    const { type, value, onChange } = props
    if (type === 'text') {
        return (
            <Input
                className="flex-1"
                placeholder={props.placeholder}
                value={value}
                onChange={event => onChange?.(event.target.value)}
            />
        )
    }
    if (type === 'select') {
        return (
            <Select
                aria-labelledby="Select filter"
                className="m-0 p-0"
                value={value}
                isCustomStyle={true}
                onChange={value => onChange?.(value.target.value)}
            >
                {props.options.map(({ label, value }) => (
                    <option key={label} value={value}>
                        {label}
                    </option>
                ))}
            </Select>
        )
    }
    if (type === 'date-range') {
        return <DateRangeSelect {...props} value={value} onChange={onChange} />
    }
    return null
}

export interface IColumn<T> {
    key: string
    accessor?: keyof T | ((data: T) => any)
    header:
        | string
        | {
              label: string
              align: 'left' | 'right' | 'center'
              tooltip?: string
          }
    sortable?: boolean
    align?: 'left' | 'right' | 'center'
    render?: (data: T, index: number) => JSX.Element | null
    filter?: ColumnFilterProps
    cellClassName?: ((data: T, index: number) => string) | string
}

interface IAction<T> {
    key: string
    label: string
    icon: string
    iconColor?: 'muted' | 'danger'
    labelColor?: 'body' | 'danger'
    onClick?: (items: T[]) => void
    bulk?: boolean
    condition?: (items: T[]) => boolean
    href?: (items: T[]) => string
    target?: '_blank'
}

interface TableProps<T> {
    columns: IColumn<T>[]
    data: T[]
    actions?: IAction<T>[]
    selectable?: boolean
    note?: string | JSX.Element
    getRowId: (data: T) => string
    sortBy?: {
        key: string
        descending: boolean
    }
    onClearAllFiltersClick?: () => void
    onSortByChange?: (newOderBy: NonNullable<TableProps<T>['sortBy']>) => void
    pagination?: PaginationProps
    rowClassName?: ((data: T) => string) | string
}

export function Table<T>({
    data,
    columns,
    actions = [],
    note,
    sortBy,
    getRowId,
    onSortByChange,
    selectable = false,
    onClearAllFiltersClick,
    pagination,
    rowClassName,
}: TableProps<T>): JSX.Element {
    const [selection, setSelection] = useState<T[]>([])

    useEffect(() => {
        setSelection([])
    }, [data])

    const isAllSelected = useMemo(() => {
        const selectedIDs = selection.map(getRowId)
        const allIDs = data.map(getRowId)
        return isEqual(selectedIDs.sort(), allIDs.sort()) && selectedIDs.length > 0
    }, [data, getRowId, selection])

    const onRowSelectionChange = useCallback(
        (row: T, selected: boolean): void => {
            setSelection(selection => [
                ...selection.filter(selectedRow => getRowId(selectedRow) !== getRowId(row)),
                ...(selected ? [row] : []),
            ])
        },
        [getRowId]
    )

    const bulkActions = useMemo(() => actions.filter(action => action.bulk), [actions])

    const memoizedColumns = useMemo((): IColumn<T>[] => {
        const allColumns = [...columns]

        if (actions.length > 0) {
            allColumns.push({
                key: 'actions',
                header: { label: 'Actions', align: 'right' },
                align: 'right',
                render: function RenderActions(user: T) {
                    return (
                        <div className="d-flex justify-content-end">
                            <Actions actions={actions} selection={[user]} className="border-0">
                                ...
                            </Actions>
                        </div>
                    )
                },
            })
        }

        return allColumns
    }, [actions, columns])

    const onPreviousPage = useCallback(() => {
        if (pagination) {
            pagination.onPrevious()
            setSelection([])
        }
    }, [pagination])

    const onNextPage = useCallback(() => {
        if (pagination) {
            pagination.onNext()
            setSelection([])
        }
    }, [pagination])

    const onLimitChange = useCallback(
        (newLimit: number) => {
            if (pagination?.onLimitChange) {
                pagination.onLimitChange(newLimit)
                setSelection([])
            }
        },
        [pagination]
    )

    return (
        <>
            {(selectable || pagination) && (
                <div className="mb-4 d-flex justify-content-between">
                    {selectable && <SelectionActions<T> actions={bulkActions} position="top" selection={selection} />}
                    {pagination && (
                        <Pagination
                            {...pagination}
                            onPrevious={onPreviousPage}
                            onLimitChange={onLimitChange}
                            onNext={onNextPage}
                        />
                    )}
                </div>
            )}
            <table className={styles.table}>
                <thead>
                    <tr>
                        {selectable && (
                            <th>
                                <div className={classNames(styles.header, styles.selectionHeader)}>
                                    <Checkbox
                                        aria-labelledby="Select all checkbox"
                                        className="m-0"
                                        checked={isAllSelected}
                                        onChange={event => {
                                            if (event.target.checked) {
                                                setSelection(data)
                                            } else {
                                                setSelection([] as T[])
                                            }
                                        }}
                                    />
                                </div>
                            </th>
                        )}
                        {memoizedColumns.map(column => {
                            const key = column.key
                            const label = typeof column.header === 'string' ? column.header : column.header.label
                            const align = typeof column.header !== 'string' ? column.header.align || 'left' : 'left'
                            const tooltip = typeof column.header !== 'string' ? column.header.tooltip : undefined

                            const handleSort = (): void => {
                                onSortByChange?.({ key, descending: sortBy?.key === key && !sortBy?.descending })
                            }
                            return (
                                <th key={key} onClick={column.sortable ? handleSort : undefined}>
                                    <div
                                        className={classNames('text-nowrap', styles.header, {
                                            [styles.sortedAsc]: sortBy?.key === key && !sortBy.descending,
                                            [styles.sortedDesc]: sortBy?.key === key && sortBy.descending,
                                            [styles.sortable]: column.sortable,
                                            [styles.alignRight]: align === 'right',
                                            [styles.alignCenter]: align === 'center',
                                        })}
                                    >
                                        <Tooltip content={tooltip}>
                                            <Text as="span" weight="bold">
                                                {label}
                                                {tooltip && <span className={styles.linkColor}>*</span>}
                                            </Text>
                                        </Tooltip>
                                        {column.sortable && (
                                            <div className={classNames('d-flex flex-column', styles.sortableIcons)}>
                                                <Icon
                                                    aria-label="Sort ascending"
                                                    svgPath={mdiMenuUp}
                                                    className={styles.sortAscIcon}
                                                />
                                                <Icon
                                                    aria-label="Sort descending"
                                                    svgPath={mdiMenuDown}
                                                    className={styles.sortDescIcon}
                                                />
                                            </div>
                                        )}
                                    </div>
                                </th>
                            )
                        })}
                    </tr>
                    <tr>
                        {selectable && <th />}
                        {columns.map(({ key, filter }) => (
                            <th key={key} className="pr-2">
                                {filter && <ColumnFilter {...filter} />}
                            </th>
                        ))}
                        <th>
                            {onClearAllFiltersClick && (
                                <Tooltip content="Clear filters">
                                    <Button onClick={onClearAllFiltersClick} className="text-right" display="block">
                                        <Icon aria-label="Clear filters" svgPath={mdiFilterRemoveOutline} />
                                    </Button>
                                </Tooltip>
                            )}
                        </th>
                    </tr>
                </thead>
                <tbody>
                    {data.map(item => (
                        <Row
                            key={getRowId(item)}
                            data={item}
                            columns={memoizedColumns}
                            selectable={selectable}
                            selection={selection}
                            getRowId={getRowId}
                            onSelectionChange={onRowSelectionChange}
                            rowClassName={rowClassName}
                        />
                    ))}
                </tbody>
            </table>
            {selectable && (
                <div className="mt-4 d-flex justify-content-between align-items-center">
                    <SelectionActions<T> actions={bulkActions} position="bottom" selection={selection} />
                    {note}
                </div>
            )}
        </>
    )
}

interface RowProps<T> {
    data: T
    columns: IColumn<T>[]
    selectable: boolean
    selection: T[]
    getRowId: (data: T) => string | number
    onSelectionChange: (data: T, selected: boolean) => void
    rowClassName?: ((data: T) => string) | string
}

function Row<T>({
    data,
    columns,
    selectable,
    selection,
    getRowId,
    onSelectionChange,
    rowClassName,
}: RowProps<T>): JSX.Element {
    const rowKey = getRowId(data)
    const isSelected = useMemo(() => !!selection.find(row => getRowId(row) === rowKey), [getRowId, rowKey, selection])

    return (
        <tr className={typeof rowClassName === 'function' ? rowClassName(data) : rowClassName}>
            {selectable && (
                <td className={styles.selectionTd}>
                    <div className={classNames(styles.cell, styles.selection)}>
                        <Checkbox
                            aria-labelledby={`${rowKey} selection checkbox`}
                            className="m-0"
                            checked={isSelected}
                            onChange={event => onSelectionChange(data, event.target.checked)}
                        />
                    </div>
                </td>
            )}
            {columns.map(({ align, accessor, render, key, cellClassName }, index) => (
                <td
                    key={key}
                    className={typeof cellClassName === 'function' ? cellClassName(data, index) : cellClassName}
                >
                    {render ? (
                        render(data, index)
                    ) : (
                        <div
                            className={classNames(styles.cell, {
                                [styles.alignLeft]: !align || align === 'left',
                                [styles.alignRight]: align === 'right',
                            })}
                        >
                            <Text alignment={align || 'left'} className="mb-0">
                                {typeof accessor === 'function'
                                    ? accessor(data)
                                    : accessor !== undefined
                                    ? data[accessor]
                                    : 'n/a'}
                            </Text>
                        </div>
                    )}
                </td>
            ))}
        </tr>
    )
}

interface SelectionActionsProps<T> {
    actions: IAction<T>[]
    position: 'top' | 'bottom'
    selection: T[]
}

function SelectionActions<T>({ actions, position, selection }: SelectionActionsProps<T>): JSX.Element {
    return (
        <div className="d-flex align-items-center">
            <Icon
                svgPath={position === 'top' ? mdiArrowRightTop : mdiArrowRightBottom}
                size="md"
                aria-label={position === 'top' ? 'Sort descending' : 'Sort ascending'}
                className={classNames(styles.actionsArrowIcon, 'ml-2 mr-1')}
            />
            <Actions actions={actions} selection={selection} disabled={!selection.length}>
                {selection.length ? `With ${selection.length} selected` : 'Actions'}
                <Icon aria-label="Arrow down" svgPath={mdiChevronDown} className="ml-1" />
            </Actions>
        </div>
    )
}

interface ActionsProps<T> {
    selection: T[]
    actions: IAction<T>[]
    disabled?: boolean
    children?: React.ReactNode
    className?: string
}

function Actions<T>({ children, actions, disabled, selection, className }: ActionsProps<T>): JSX.Element {
    const [isOpen, setIsOpen] = useState<boolean>(false)
    const handleOpenChange = useCallback((event: PopoverOpenEvent): void => {
        setIsOpen(event.isOpen)
    }, [])

    const filteredActions = actions.filter(({ condition }) => !condition || condition(selection))

    return (
        <Popover isOpen={isOpen} onOpenChange={handleOpenChange}>
            <PopoverTrigger as={Button} className={className} disabled={disabled} variant="secondary" outline={true}>
                {children}
            </PopoverTrigger>
            <PopoverContent position={Position.bottom} focusLocked={false}>
                {filteredActions.length > 0 ? (
                    <ul className="list-unstyled mb-0">
                        {filteredActions.map(({ key, label, icon, iconColor, labelColor, onClick, href, target }) => (
                            <Button
                                className={styles.actionItem}
                                key={key}
                                variant="link"
                                as={href ? 'a' : 'li'}
                                href={href?.(selection)}
                                target={target}
                                outline={false}
                                onClick={() => {
                                    onClick?.(selection)
                                    setIsOpen(false)
                                }}
                            >
                                <Icon
                                    aria-label={label}
                                    svgPath={icon}
                                    size="md"
                                    className={`text-${iconColor || 'muted'}`}
                                />
                                <span className={classNames('ml-2', labelColor === 'danger' && 'text-danger')}>
                                    {label}
                                </span>
                            </Button>
                        ))}
                    </ul>
                ) : (
                    <Text className="m-2 font-italic text-muted">No actions available</Text>
                )}
            </PopoverContent>
        </Popover>
    )
}

interface PaginationProps {
    total: number
    offset: number
    limit: number
    limitOptions?: { value: number; label: string }[]
    onPrevious: () => void
    onNext: () => void
    onLimitChange: (limit: number) => void
    formatLabel?: (start: number, end: number, total: number) => string
}

const Pagination: React.FunctionComponent<PaginationProps> = ({
    onPrevious,
    limit,
    offset,
    total,
    limitOptions,
    onLimitChange,
    onNext,
    formatLabel = (start: number, end: number, total: number) => `${start}-${end} of ${total}`,
}) => {
    const [isOpen, setIsOpen] = useState<boolean>(false)
    const handleOpenChange = useCallback((event: PopoverOpenEvent): void => {
        setIsOpen(event.isOpen)
    }, [])
    return (
        <div className={classNames('d-flex justify-content-between align-items-center', styles.pagination)}>
            <Button className="mr-2" onClick={onPrevious}>
                <Icon aria-label="Show previous icon" svgPath={mdiChevronLeft} />
            </Button>
            <Popover isOpen={isOpen} onOpenChange={handleOpenChange}>
                <PopoverTrigger as={Text} className="m-0 p-0 cursor-pointer">
                    <Text as="span">{formatLabel(Math.min(offset + 1), Math.min(offset + limit, total), total)}</Text>
                </PopoverTrigger>
                <PopoverContent focusLocked={false}>
                    <ul className="list-unstyled mb-0">
                        {limitOptions?.map(item => (
                            <Button
                                className="d-flex cursor-pointer"
                                key={item.value}
                                variant="link"
                                as="li"
                                outline={true}
                                onClick={() => {
                                    onLimitChange(item.value)
                                    setIsOpen(false)
                                }}
                            >
                                <span
                                    className={classNames(
                                        item.value === limit ? 'font-weight-medium' : 'font-weight-normal ml-3'
                                    )}
                                >
                                    {item.value === limit && '✓'} {item.label}
                                </span>
                            </Button>
                        ))}
                    </ul>
                </PopoverContent>
            </Popover>
            <Button onClick={onNext}>
                <Icon aria-label="Show next icon" svgPath={mdiChevronRight} />
            </Button>
        </div>
    )
}
