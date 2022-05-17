import React, { useCallback, useState } from 'react'

import classNames from 'classnames'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import ChevronDownIcon from 'mdi-react/ChevronDownIcon'
import ChevronLeftIcon from 'mdi-react/ChevronLeftIcon'
import InformationOutlineIcon from 'mdi-react/InformationOutlineIcon'
import SearchIcon from 'mdi-react/SearchIcon'
// eslint-disable-next-line no-restricted-imports

import { Form } from '@sourcegraph/branded/src/components/Form'
import { renderMarkdown } from '@sourcegraph/common'
import { SyntaxHighlightedSearchQuery } from '@sourcegraph/search-ui'
import { Markdown } from '@sourcegraph/shared/src/components/Markdown'
import { Skipped } from '@sourcegraph/shared/src/search/stream'
import { Button, Collapse, CollapseHeader, CollapsePanel, Icon, Checkbox, Typography } from '@sourcegraph/wildcard'

import { StreamingProgressProps } from './StreamingProgress'

import styles from './StreamingProgressSkippedPopover.module.scss'

const severityToNumber = (severity: Skipped['severity']): number => {
    switch (severity) {
        case 'error':
            return 1
        case 'warn':
            return 2
        case 'info':
            return 3
    }
}

const sortBySeverity = (a: Skipped, b: Skipped): number => {
    const aSev = severityToNumber(a.severity)
    const bSev = severityToNumber(b.severity)

    return aSev - bSev
}

const SkippedMessage: React.FunctionComponent<React.PropsWithChildren<{ skipped: Skipped; startOpen: boolean }>> = ({
    skipped,
    startOpen,
}) => {
    // Reactstrap is preventing default behavior on all non-DropdownItem elements inside a Dropdown,
    // so we need to stop propagation to allow normal behavior (e.g. enter and space to activate buttons)
    // See Reactstrap bug: https://github.com/reactstrap/reactstrap/issues/2099
    const onKeyDown = useCallback((event: React.KeyboardEvent<HTMLButtonElement>): void => {
        if (event.key === ' ' || event.key === 'Enter') {
            event.stopPropagation()
        }
    }, [])

    return (
        <div
            className={classNames(
                'pt-2 w-100',
                styles.streamingSkippedItem,
                skipped.severity !== 'info' && styles.streamingSkippedItemWarn
            )}
        >
            <Collapse openByDefault={startOpen}>
                {({ isOpen }) => (
                    <>
                        <CollapseHeader
                            className={classNames(styles.button, 'p-2 w-100 bg-transparent border-0')}
                            onKeyDown={onKeyDown}
                            disabled={!skipped.message}
                            as={Button}
                            outline={true}
                            variant={skipped.severity !== 'info' ? 'danger' : 'primary'}
                        >
                            <Typography.H4 className="d-flex align-items-center mb-0 w-100">
                                <Icon
                                    className={classNames(styles.icon, 'flex-shrink-0')}
                                    as={skipped.severity === 'info' ? InformationOutlineIcon : AlertCircleIcon}
                                />

                                <span className="flex-grow-1 text-left">{skipped.title}</span>

                                {skipped.message && (
                                    <Icon
                                        className={classNames('flex-shrink-0', styles.chevron)}
                                        as={isOpen ? ChevronDownIcon : ChevronLeftIcon}
                                    />
                                )}
                            </Typography.H4>
                        </CollapseHeader>

                        {skipped.message && (
                            <CollapsePanel>
                                <Markdown
                                    className={classNames(styles.message, styles.markdown, 'text-left py-1')}
                                    dangerousInnerHTML={renderMarkdown(skipped.message)}
                                />
                            </CollapsePanel>
                        )}
                    </>
                )}
            </Collapse>
            <div className={classNames(styles.bottomBorderSpacer, 'mt-2')} />
        </div>
    )
}

export const StreamingProgressSkippedPopover: React.FunctionComponent<
    React.PropsWithChildren<Pick<StreamingProgressProps, 'progress' | 'onSearchAgain'>>
> = ({ progress, onSearchAgain }) => {
    const [selectedSuggestedSearches, setSelectedSuggestedSearches] = useState(new Set<string>())
    const submitHandler = useCallback(
        (event: React.FormEvent) => {
            onSearchAgain([...selectedSuggestedSearches])
            event.preventDefault()
        },
        [selectedSuggestedSearches, onSearchAgain]
    )
    const checkboxHandler = useCallback((event: React.FormEvent<HTMLInputElement>) => {
        const itemToToggle = event.currentTarget.value
        const checked = event.currentTarget.checked
        setSelectedSuggestedSearches(selected => {
            const newSelected = new Set(selected)
            if (checked) {
                newSelected.add(itemToToggle)
            } else {
                newSelected.delete(itemToToggle)
            }
            return newSelected
        })
    }, [])

    const sortedSkippedItems = progress.skipped.sort(sortBySeverity)

    return (
        <>
            {sortedSkippedItems.map((skipped, index) => (
                <SkippedMessage
                    key={skipped.reason}
                    skipped={skipped}
                    // Start with first item open, but only if it's not info severity or if there's only one item
                    startOpen={index === 0 && (skipped.severity !== 'info' || sortedSkippedItems.length === 1)}
                />
            ))}
            {sortedSkippedItems.some(skipped => skipped.suggested) && (
                <Form className="pb-3 px-3" onSubmit={submitHandler} data-testid="popover-form">
                    <div className="mb-2 mt-3">Search again:</div>
                    <div className="form-check">
                        {sortedSkippedItems.map(
                            (skipped, index) =>
                                skipped.suggested && (
                                    <Checkbox
                                        key={skipped.suggested.queryExpression}
                                        value={skipped.suggested.queryExpression}
                                        onChange={checkboxHandler}
                                        data-testid={`streaming-progress-skipped-suggest-check-${index}`}
                                        id={`streaming-progress-skipped-suggest-check-${index}`}
                                        wrapperClassName="mb-1 d-block"
                                        label={
                                            <>
                                                {skipped.suggested.title} (
                                                <SyntaxHighlightedSearchQuery
                                                    query={skipped.suggested.queryExpression}
                                                />
                                                )
                                            </>
                                        }
                                    />
                                )
                        )}
                    </div>

                    <Button
                        type="submit"
                        className="mt-2"
                        variant="primary"
                        disabled={selectedSuggestedSearches.size === 0}
                        data-testid="skipped-popover-form-submit-btn"
                    >
                        <Icon className="mr-1" as={SearchIcon} />
                        Search again
                    </Button>
                </Form>
            )}
        </>
    )
}
