import { useState, useMemo } from 'react'
import classNames from 'classnames'
import { mdiMenuUp, mdiMenuDown, mdiArrowRightTop, mdiArrowRightBottom, mdiChevronDown } from '@mdi/js'
import {
    Icon,
    H2,
    Text,
    Checkbox,
    PopoverTrigger,
    PopoverContent,
    Popover,
    Position,
    Button,
} from '@sourcegraph/wildcard'

import styles from './index.module.scss'

interface IColumn<TData> {
    key: string
    accessor?: string | ((data: TData) => any)
    header:
        | string
        | {
              label: string
              align: 'left' | 'right'
          }
    sortable?: boolean
    align?: 'left' | 'right' | 'center'
    render?: (data: TData, index: number) => JSX.Element
}

interface IAction<TData> {
    key: string
    label: string
    icon: string
    iconColor?: 'muted' | 'danger'
    labelColor?: 'body' | 'danger'
    onClick: (ids: (string | number)[]) => void
}

type GetRowId = <TData>(data: TData) => string | number

interface IProps<TData extends object> {
    columns: IColumn<TData>[]
    data: TData[]
    actions?: IAction<TData>[]
    selectable?: boolean
    note?: string | JSX.Element
    getRowId?: GetRowId<TData>
    initialSortColumn?: string
    initialSortDirection?: 'asc' | 'desc'
    onSortChange?: (column: string | null, direction: 'asc' | 'desc' | null) => void
    onSelectionChange?: (rows: TData[]) => void
}

type ISelectionState = { [key: string | number]: boolean }

export default function Table<TData>({
    data,
    columns,
    selectable = false,
    actions = [],
    note,
    getRowId = (data: any) => data.id,
    onSortChange,
    initialSortColumn = null,
    initialSortDirection = 'asc',
}: IProps<TData>) {
    const [sortedColumn, setSortedColumn] = useState<string | null>(initialSortColumn)
    const [sortDirection, setSortDirection] = useState<'asc' | 'desc' | null>(initialSortDirection)
    const [selection, setSelection] = useState<TData[]>([])

    const onRowSelectionChange = (row, selected) => {
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

                            return (
                                <th
                                    key={key}
                                    onClick={
                                        column.sortable &&
                                        (() => {
                                            let newColumn = sortedColumn
                                            let newDirection = sortDirection

                                            if (sortedColumn !== key) {
                                                newColumn = key
                                                newDirection = 'asc'
                                            } else if (sortDirection === 'desc') {
                                                newColumn = null
                                                newDirection = null
                                            } else {
                                                newDirection = 'desc'
                                            }
                                            setSortedColumn(newColumn)
                                            setSortDirection(newDirection)

                                            onSortChange?.(newColumn, newDirection)
                                        })
                                    }
                                >
                                    <div
                                        className={classNames(styles.header, styles.sortable, {
                                            [styles.alignRight]: align === 'right',
                                            [styles.sortedAsc]: sortedColumn === key && sortDirection === 'asc',
                                            [styles.sortedDesc]: sortedColumn === key && sortDirection === 'desc',
                                        })}
                                    >
                                        <Text as="span" weight="bold">
                                            {label}
                                        </Text>
                                        {column.sortable && (
                                            <div className={classNames('d-flex flex-column', styles.sortableIcons)}>
                                                <Icon svgPath={mdiMenuUp} size="md" className={styles.sortAscIcon} />
                                                <Icon svgPath={mdiMenuDown} size="md" className={styles.sortDescIcon} />
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

function Row<TData>({
    data,
    columns,
    selectable,
    selection,
    getRowId,
    onSelectionChange,
}: {
    data: TData
    columns: IColumn<TData>[]
    selectable: boolean
    selection: TData[]
    getRowId: GetRowId<TData>
    onSelectionChange: (data: TData, selected: boolean) => void
}) {
    const key = getRowId(data)
    const selected = useMemo(() => {
        !!selection.find(row => getRowId(row) === key)
    }, selection)

    return (
        <tr key={key}>
            {selectable && (
                <td className={styles.selectionTd}>
                    <div className={classNames(styles.cell, styles.selection)}>
                        <Checkbox
                            className="m-0"
                            checked={selected}
                            onChange={e => onSelectionChange(data, e.target.checked)}
                        />
                    </div>
                </td>
            )}
            {columns.map(column => (
                <td key={columns.key}>
                    {!!column.render ? (
                        column.render(data)
                    ) : (
                        <div className={styles.cell}>
                            <Text alignment={column.align || 'left'} className="mb-0">
                                {typeof column.accessor === 'function'
                                    ? column.accessor(data)
                                    : typeof column.accessor !== 'undefined'
                                    ? data[column.accessor]
                                    : ''}
                            </Text>
                        </div>
                    )}
                </td>
            ))}
        </tr>
    )
}

function SelectionActions<TData>({
    actions,
    position,
    selection,
}: {
    actions: IAction<TData>[]
    position: 'top' | 'bottom'
    selection: TData[]
}) {
    return (
        <div className="d-flex align-items-center">
            <Icon
                svgPath={position === 'top' ? mdiArrowRightTop : mdiArrowRightBottom}
                size="md"
                className={styles.actionsArrowIcon}
            />
            <Text className="mx-2 my-0">
                {!!selection.length ? `With ${selection.length} selected` : 'With selected'}
            </Text>
            <Popover>
                <PopoverTrigger as={Button} variant="secondary" disabled={!selection.length}>
                    Actions
                    <Icon svgPath={mdiChevronDown} className="ml-1" />
                </PopoverTrigger>
                <PopoverContent position={Position.bottom}>
                    <ul className="list-unstyled mb-0">
                        {actions.map(action => (
                            <li key={action.key} className="d-flex p-2">
                                <Icon
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
