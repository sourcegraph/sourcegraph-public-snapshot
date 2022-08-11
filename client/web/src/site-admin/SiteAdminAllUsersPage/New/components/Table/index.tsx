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

interface IAction<TData> {
    key: string
    label: string
    icon: string
    iconColor?: 'muted' | 'danger'
    labelColor?: 'body' | 'danger'
    onClick: (items: TData[]) => void
}

interface TableProps<TData> {
    columns: IColumn<TData>[]
    data: TData[]
    actions?: IAction<TData>[]
    selectable?: boolean
    note?: string | JSX.Element
    getRowId?: (data: TData) => string | number
    sortBy?: {
        key: string
        descending?: boolean
    }
    onSortByChange?: (newOderBy: NonNullable<TableProps<TData>['sortBy']>) => void
    onSelectionChange?: (rows: TData[]) => void
}

export function Table<TData>({
    data,
    columns,
    selectable = false,
    actions = [],
    note,
    getRowId = (data: any) => data.id,
    onSortByChange,
    sortBy,
    onSelectionChange,
}: TableProps<TData>): JSX.Element {
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

                            const handleSort = (): void => {
                                onSortByChange?.({ key, descending: sortBy?.key === key && !sortBy?.descending })
                            }
                            return (
                                <th key={key} onClick={column.sortable ? handleSort : undefined}>
                                    <div
                                        className={classNames(
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
    const rowKey = getRowId(data)
    const selected = useMemo(() => !!selection.find(row => getRowId(row) === rowKey), [getRowId, rowKey, selection])

    return (
        <tr>
            {selectable && (
                <td className={styles.selectionTd}>
                    <div className={classNames(styles.cell, styles.selection)}>
                        <Checkbox
                            aria-labelledby={`${rowKey} selection checkbox`}
                            className="m-0"
                            checked={selected}
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

interface SelectionActionsProps<TData> {
    actions: IAction<TData>[]
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
                        {actions.map(({ key, label, icon, iconColor, labelColor, onClick }) => (
                            <Button
                                className="d-flex cursor-pointer"
                                key={key}
                                variant="link"
                                as="li"
                                outline={true}
                                onClick={() => onClick?.(selection)}
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
        </div>
    )
}
