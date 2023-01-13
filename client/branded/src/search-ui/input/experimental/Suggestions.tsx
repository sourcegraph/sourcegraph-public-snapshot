import React, { MouseEvent, useMemo, useState, useCallback } from 'react'

import { Icon, useWindowSize } from '@sourcegraph/wildcard'

import { SyntaxHighlightedSearchQuery } from '../../components'

import { Group, Option } from './suggestionsExtension'

import styles from './Suggestions.module.scss'

function getNote(option: Option): string {
    switch (option.type) {
        case 'completion':
            return 'Add'
        case 'target':
            return option.note ?? 'Jump to'
        case 'command':
            return option.note ?? ''
    }
}

interface SuggesionsProps {
    id: string
    results: Group[]
    activeRowIndex: number
    open?: boolean
    onSelect(option: Option): void
}

export const Suggestions: React.FunctionComponent<SuggesionsProps> = ({
    id,
    results,
    activeRowIndex,
    onSelect,
    open = false,
}) => {
    const [container, setContainer] = useState<HTMLDivElement | null>(null)

    const handleSelection = useCallback(
        (event: MouseEvent) => {
            const match = (event.target as HTMLElement).closest('li[role="row"]')?.id.match(/\d+x\d+/)
            if (match) {
                // Extracts the group and row index from the elements ID to pass
                // the right option value to the callback.
                const [group, option] = match[0].split('x')
                onSelect(results[+group].options[+option])
            }
        },
        [onSelect, results]
    )

    const { height: windowHeight } = useWindowSize()
    const maxHeight = useMemo(
        // This is using an arbitrary 20px "margin" between the suggestions box
        // and the window border
        () => (container ? `${windowHeight - container.getBoundingClientRect().top - 20}px` : 'auto'),
        [container, windowHeight]
    )
    const flattenedRows = useMemo(() => results.flatMap(group => group.options), [results])
    const focusedItem = flattenedRows[activeRowIndex]
    const show = open && results.length > 0

    if (!show) {
        return null
    }

    return (
        <div
            ref={setContainer}
            id={id}
            className={styles.suggestions}
            role="grid"
            // eslint-disable-next-line react/forbid-dom-props
            style={{ maxHeight }}
            onMouseDown={handleSelection}
            tabIndex={-1}
        >
            {results.map((group, groupIndex) =>
                group.options.length > 0 ? (
                    <ul role="rowgroup" key={group.title} aria-labelledby={`${id}-${groupIndex}-label`}>
                        <li id={`${id}-${groupIndex}-label`} role="presentation">
                            {group.title}
                        </li>
                        {group.options.map((option, rowIndex) => (
                            <li
                                role="row"
                                key={rowIndex}
                                id={`${id}-${groupIndex}x${rowIndex}`}
                                aria-selected={focusedItem === option}
                            >
                                {option.icon && (
                                    <div className="pr-1">
                                        <Icon className={styles.icon} svgPath={option.icon} aria-hidden="true" />
                                    </div>
                                )}
                                <div role="gridcell">
                                    {option.render
                                        ? option.render(option)
                                        : option.matches
                                        ? [...option.value].map((char, index) =>
                                              option.matches!.has(index) ? (
                                                  <span key={index} className={styles.match}>
                                                      {char}
                                                  </span>
                                              ) : (
                                                  char
                                              )
                                          )
                                        : option.value}
                                </div>
                                {option.description && (
                                    <div role="gridcell" className={styles.description}>
                                        {option.description}
                                    </div>
                                )}
                                <div role="gridcell" className={styles.note}>
                                    {getNote(option)}
                                </div>
                            </li>
                        ))}
                    </ul>
                ) : null
            )}
        </div>
    )
}

export const FilterOption: React.FunctionComponent<{ option: Option }> = ({ option }) => (
    <span className={styles.filterOption}>
        {option.matches
            ? [...option.value].map((char, index) =>
                  option.matches!.has(index) ? (
                      <span key={index} className={styles.match}>
                          {char}
                      </span>
                  ) : (
                      char
                  )
              )
            : option.value}
        <span className={styles.separator}>:</span>
    </span>
)

export const QueryOption: React.FunctionComponent<{ option: Option }> = ({ option }) => (
    <SyntaxHighlightedSearchQuery query={option.value} />
)
