import { useState, useMemo } from 'react'

import { mdiMenuUp, mdiMenuDown, mdiArrowRightTop, mdiArrowRightBottom, mdiChevronDown } from '@mdi/js'
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
} from '@sourcegraph/wildcard'

import styles from './index.module.scss'

interface IColumn<TData> {
    key: string
    accessor?: keyof TData | ((data: TData) => any)
    header:
        | string
        | {
              label: string
              align: 'left' | 'right'
              tooltip?: string
          }
    sortable?: boolean
    align?: 'left' | 'right' | 'center'
    render?: (data: TData, index: number) => JSX.Element
}

interface IAction {
    key: string
    label: string
    icon: string
    iconColor?: 'muted' | 'danger'
    labelColor?: 'body' | 'danger'
    onClick: (ids: (string | number)[]) => void
}

interface TableProps<TData> {
    columns: IColumn<TData>[]
    data: TData[]
    actions?: IAction[]
    selectable?: boolean
    note?: string | JSX.Element
    getRowId?: (data: TData) => string | number
    initialSortColumn?: string
    initialSortDirection?: 'asc' | 'desc'
    onSortChange?: (column: string | undefined, direction: 'asc' | 'desc' | undefined) => void
    onSelectionChange?: (rows: TData[]) => void
}

export function Table<TData>({
    data,
    columns,
    selectable = false,
    actions = [],
    note,
    getRowId = (data: any) => data.id,
    onSortChange,
    initialSortColumn,
    initialSortDirection = 'asc',
    onSelectionChange,
}: TableProps<TData>): JSX.Element {
    const [sortedColumn, setSortedColumn] = useState(initialSortColumn)
    const [sortDirection, setSortDirection] = useState<'asc' | 'desc' | null>(initialSortDirection)
    const [selection, setSelection] = useState<TData[]>([])

    const onRowSelectionChange = (row: TData, selected: boolean): void => {
        const newSelection = selection.filter(selectedRow => getRowId(selectedRow) !== getRowId(row))
        if (selected) {
            newSelection.push(row)
        }

        setSelection(newSelection)
        onSelectionChange?.(newSelection)
    }

    return (
        <>
            {selectable && (
                <div className="mb-4">
                    <SelectionActions<TData> actions={actions} position="top" selection={selection} />
                </div>
            )}
            <table className={styles.table}>
                <thead>
                    <tr>
                        {selectable && (
                            <th>
                                <div className={styles.header} />
                            </th>
                        )}
                        {columns.map(column => {
                            const key = column.key
                            const label = typeof column.header === 'string' ? column.header : column.header.label
                            const align = typeof column.header !== 'string' ? column.header.align || 'left' : 'left'
                            const tooltip = typeof column.header !== 'string' ? column.header.tooltip : undefined

                            const sort = (): void => {
                                let newColumn = sortedColumn
                                let newDirection = sortDirection

                                if (sortedColumn !== key) {
                                    newColumn = key
                                    newDirection = 'asc'
                                } else if (sortDirection === 'desc') {
                                    newColumn = undefined
                                    newDirection = undefined
                                } else {
                                    newDirection = 'desc'
                                }
                                setSortedColumn(newColumn)
                                setSortDirection(newDirection)

                                onSortChange?.(newColumn, newDirection)
                            }
                            return (
                                <th key={key} onClick={column.sortable ? sort : undefined}>
                                    <div
                                        className={classNames(styles.header, styles.sortable, {
                                            [styles.alignRight]: align === 'right',
                                            [styles.sortedAsc]: sortedColumn === key && sortDirection === 'asc',
                                            [styles.sortedDesc]: sortedColumn === key && sortDirection === 'desc',
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
                                                    size="md"
                                                    className={styles.sortAscIcon}
                                                />
                                                <Icon
                                                    aria-label="Sort descending"
                                                    svgPath={mdiMenuDown}
                                                    size="md"
                                                    className={styles.sortDescIcon}
                                                />
                                            </div>
                                        )}
                                    </div>
                                </th>
                            )
                        })}
                    </tr>
                </thead>
                <tbody>
                    {data.map(item => (
                        <Row<TData>
                            key={getRowId(item)}
                            data={item}
                            columns={columns}
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
                    <SelectionActions<TData> actions={actions} position="bottom" selection={selection} />
                    {note}
                </div>
            )}
        </>
    )
}

interface RowProps<TData> {
    data: TData
    columns: IColumn<TData>[]
    selectable: boolean
    selection: TData[]
    getRowId: (data: TData) => string | number
    onSelectionChange: (data: TData, selected: boolean) => void
}

function Row<TData>({
    data,
    columns,
    selectable,
    selection,
    getRowId,
    onSelectionChange,
}: RowProps<TData>): JSX.Element {
    const key = getRowId(data)
    const selected = useMemo(() => !!selection.find(row => getRowId(row) === key), [getRowId, key, selection])

    return (
        <tr key={key}>
            {selectable && (
                <td className={styles.selectionTd}>
                    <div className={classNames(styles.cell, styles.selection)}>
                        <Checkbox
                            aria-labelledby={`${key} selection checkbox`}
                            className="m-0"
                            checked={selected}
                            onChange={event => onSelectionChange(data, event.target.checked)}
                        />
                    </div>
                </td>
            )}
            {columns.map(({ align, accessor, render }, index) => (
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

interface SelectionActionsProps<TData> {
    actions: IAction[]
    position: 'top' | 'bottom'
    selection: TData[]
}

function SelectionActions<TData>({ actions, position, selection }: SelectionActionsProps<TData>): JSX.Element {
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
            <Popover>
                <PopoverTrigger as={Button} disabled={!selection.length} variant="secondary" outline={true}>
                    Actions
                    <Icon aria-label="Arrow down" svgPath={mdiChevronDown} className="ml-1" />
                </PopoverTrigger>
                <PopoverContent position={Position.bottom}>
                    <ul className="list-unstyled mb-0">
                        {actions.map(action => (
                            <li key={action.key} className="d-flex p-2">
                                <Icon
                                    aria-label={action.label}
                                    svgPath={action.icon}
                                    size="md"
                                    className={`text-${action.iconColor || 'muted'}`}
                                />
                                <span
                                    className={classNames('ml-2', {
                                        'text-danger': action.labelColor === 'danger',
                                    })}
                                >
                                    {action.label}
                                </span>
                            </li>
                        ))}
                    </ul>
                </PopoverContent>
            </Popover>
        </div>
    )
}
