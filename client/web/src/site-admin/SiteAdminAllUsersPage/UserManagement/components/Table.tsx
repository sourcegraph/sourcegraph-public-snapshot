import React, { useState, useMemo, useEffect, useCallback } from 'react'

import { mdiMenuUp, mdiMenuDown, mdiArrowRightTop, mdiArrowRightBottom, mdiChevronDown, mdiPencil } from '@mdi/js'
import classNames from 'classnames'

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
    PopoverOpenEvent,
    Input,
    Select,
} from '@sourcegraph/wildcard'

import { DateRangeSelect } from './DateRangeSelect'

import styles from './Table.module.scss'

type ColumnFilterProps =
    | {
          type: 'text'
          placeholder: string
          value?: string
          onChange?: (value: string) => void
      }
    | {
          type: 'select'
          options: { label: string; value: string }[]
          value?: string
          onChange?: (value: string) => void
      }
    | {
          type: 'date-range'
          placeholder: string
          value?: [Date, Date] | null
          /**
           * If provided, will allow "null" value selection
           */
          nullLabel?: string
          onChange?: (value?: [Date, Date] | null) => void
      }

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
        return (
            <DateRangeSelect
                placeholder={props.placeholder}
                nullLabel={props.nullLabel}
                value={value}
                onChange={onChange}
            />
        )
    }
    return null
}

interface IColumn<T> {
    key: string
    accessor?: keyof T | ((data: T) => any)
    header:
        | string
        | {
              label: string
              align: 'left' | 'right'
              tooltip?: string
          }
    sortable?: boolean
    align?: 'left' | 'right' | 'center'
    render?: (data: T, index: number) => JSX.Element
    filter?: ColumnFilterProps
}

interface IAction<T> {
    key: string
    label: string
    icon: string
    iconColor?: 'muted' | 'danger'
    labelColor?: 'body' | 'danger'
    onClick: (items: T[]) => void
    bulk?: boolean
    condition?: (items: T[]) => boolean
}

interface TableProps<T> {
    columns: IColumn<T>[]
    data: T[]
    actions?: IAction<T>[]
    selectable?: boolean
    note?: string | JSX.Element
    getRowId: (data: T) => string
    onSortByChange?: (newOderBy: NonNullable<TableProps<T>['sortBy']>) => void
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
}: TableProps<T>): JSX.Element {
    const [allSelected, setAllSelected] = useState(false)
    const [selection, setSelection] = useState<T[]>([])

    useEffect(() => {
        if (allSelected) {
            setSelection(data)
        }
    }, [allSelected, data])

    const onRowSelectionChange = useCallback(
        (row: T, selected: boolean): void => {
            setSelection(selection => [
                ...selection.filter(selectedRow => getRowId(selectedRow) !== getRowId(row)),
                ...(selected ? [row] : []),
            ])

            if (!selected) {
                setAllSelected(false)
            }
        },
        [getRowId]
    )

    const bulkActions = useMemo(() => actions.filter(action => action.bulk), [actions])

    const memoizedColumns = useMemo(
        (): IColumn<T>[] => [
            ...columns,
            {
                key: 'actions',
                header: { label: 'Actions', align: 'right' },
                align: 'right',
                render: function RenderActions(user: T) {
                    return (
                        <div className="d-flex justify-content-end">
                            <Actions actions={actions} selection={[user]}>
                                <Icon aria-label="Pencil icon" svgPath={mdiPencil} className="ml-1" />
                                <Icon aria-label="Arrow down" svgPath={mdiChevronDown} className="ml-1" />
                            </Actions>
                        </div>
                    )
                },
            },
        ],
        [actions, columns]
    )

    return (
        <>
            {selectable && (
                <div className="mb-4">
                    <SelectionActions<T> actions={bulkActions} position="top" selection={selection} />
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
                                        checked={allSelected}
                                        onChange={event => {
                                            if (event.target.checked) {
                                                setAllSelected(true)
                                                setSelection(data)
                                            } else {
                                                setAllSelected(false)
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
                                        className={classNames(
                                            'text-nowrap',
                                            styles.header,
                                            styles.sortable,
                                            align === 'right' && styles.alignRight,
                                            {
                                                [styles.sortedAsc]: sortBy?.key === key && !sortBy.descending,
                                                [styles.sortedDesc]: sortBy?.key === key && sortBy.descending,
                                            }
                                        )}
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
                        {memoizedColumns.map(({ key, filter }) => (
                            <th key={key}>{filter && <ColumnFilter {...filter} />}</th>
                        ))}
                    </tr>
                </thead>
                <tbody>
                    {data.map(item => (
                        <Row<T>
                            key={getRowId(item)}
                            data={item}
                            columns={memoizedColumns}
                            selectable={selectable}
                            selection={selection}
                            getRowId={getRowId}
                            onSelectionChange={onRowSelectionChange}
                        />
                    ))}
                </tbody>
            </table>
            {selectable && (
                <div className="mt-4 d-flex justify-content-between align-items-center">
                    <SelectionActions<T> actions={actions} position="bottom" selection={selection} />
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
}

function Row<T>({ data, columns, selectable, selection, getRowId, onSelectionChange }: RowProps<T>): JSX.Element {
    const rowKey = getRowId(data)
    const isSelected = useMemo(() => !!selection.find(row => getRowId(row) === rowKey), [getRowId, rowKey, selection])

    return (
        <tr>
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
            {columns.map(({ align, accessor, render, key }, index) => (
                <td key={key}>
                    {render ? (
                        render(data, index)
                    ) : (
                        <div className={styles.cell}>
                            <Text alignment={align || 'left'} className="mb-0">
                                {typeof accessor === 'function'
                                    ? accessor(data)
                                    : typeof accessor !== 'undefined'
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
                className={styles.actionsArrowIcon}
            />
            <Text className="mx-2 my-0">
                {selection.length ? `With ${selection.length} selected` : 'With selected'}
            </Text>
            <Actions actions={actions} selection={selection} disabled={!selection.length}>
                Actions
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
}

function Actions<T>({ children, actions, disabled, selection }: ActionsProps<T>): JSX.Element {
    const [isOpen, setIsOpen] = useState<boolean>(false)
    const handleOpenChange = useCallback((event: PopoverOpenEvent): void => {
        setIsOpen(event.isOpen)
    }, [])

    return (
        <Popover isOpen={isOpen} onOpenChange={handleOpenChange}>
            <PopoverTrigger as={Button} disabled={disabled} variant="secondary" outline={true}>
                {children}
            </PopoverTrigger>
            <PopoverContent position={Position.bottom}>
                <ul className="list-unstyled mb-0">
                    {actions
                        .filter(({ condition }) => !condition || condition(selection))
                        .map(({ key, label, icon, iconColor, labelColor, onClick }) => (
                            <Button
                                className="d-flex cursor-pointer"
                                key={key}
                                variant="link"
                                as="li"
                                outline={true}
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
            </PopoverContent>
        </Popover>
    )
}
