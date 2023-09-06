import React, { useCallback, useState, FC, useMemo } from 'react'

import { mdiAlertCircle, mdiChevronDown, mdiChevronLeft, mdiInformationOutline, mdiMagnify } from '@mdi/js'
import classNames from 'classnames'

import { pluralize, renderMarkdown } from '@sourcegraph/common'
import type { Skipped } from '@sourcegraph/shared/src/search/stream'
import { Progress } from '@sourcegraph/shared/src/search/stream'
import { useExperimentalFeatures } from '@sourcegraph/shared/src/settings/settings'
import {
    Button,
    Collapse,
    CollapseHeader,
    CollapsePanel,
    Icon,
    Checkbox,
    H4,
    Text,
    H3,
    Markdown,
    Form,
    ProductStatusBadge,
    Alert,
} from '@sourcegraph/wildcard'

import { SyntaxHighlightedSearchQuery } from '../../components'

import { validateQueryForExhaustiveSearch } from './exhaustive-search/exhaustive-search-validation'
import { sortBySeverity, limitHit } from './utils'

import styles from './StreamingProgressSkippedPopover.module.scss'

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
                            <H4 className="d-flex align-items-center mb-0 w-100">
                                <Icon
                                    aria-label={skipped.severity === 'info' ? 'Information' : 'Alert'}
                                    className={classNames(styles.icon, 'flex-shrink-0')}
                                    svgPath={skipped.severity === 'info' ? mdiInformationOutline : mdiAlertCircle}
                                />

                                <span className="flex-grow-1 text-left">{skipped.title}</span>

                                {skipped.message && (
                                    <Icon
                                        aria-hidden={true}
                                        className={classNames('flex-shrink-0', styles.chevron)}
                                        svgPath={isOpen ? mdiChevronDown : mdiChevronLeft}
                                    />
                                )}
                            </H4>
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

interface StreamingProgressSkippedPopoverProps {
    query: string
    progress: Progress
    onSearchAgain: (additionalFilters: string[]) => void
}

export const StreamingProgressSkippedPopover: FC<StreamingProgressSkippedPopoverProps> = props => {
    const { query, progress, onSearchAgain } = props

    const exhaustiveSearch = useExperimentalFeatures(settings => settings.searchJobs)
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

    const sortedSkippedItems = sortBySeverity(progress.skipped)

    return (
        <>
            <Text className={classNames('mx-3 mt-3', exhaustiveSearch && 'mb-0')}>
                Found {limitHit(progress) ? 'more than ' : ''}
                {progress.matchCount} {pluralize('result', progress.matchCount)}
                {progress.repositoriesCount !== undefined
                    ? ` from ${progress.repositoriesCount} ${pluralize(
                          'repository',
                          progress.repositoriesCount,
                          'repositories'
                      )}`
                    : ''}
                .
            </Text>

            {exhaustiveSearch ? (
                <SlimSkippedReasons items={sortedSkippedItems} />
            ) : (
                <SkippedReasons items={sortedSkippedItems} />
            )}

            {sortedSkippedItems.some(skipped => skipped.suggested) && (
                <SkippedItemsSearch
                    slim={exhaustiveSearch ?? false}
                    items={sortedSkippedItems}
                    disabled={selectedSuggestedSearches.size === 0}
                    onSearchSettingsChange={checkboxHandler}
                    onSubmit={submitHandler}
                />
            )}

            {exhaustiveSearch && (
                <>
                    <hr className="mx-3" />
                    <ExhaustiveSearchMessage query={query} />
                </>
            )}
        </>
    )
}

interface SkippedReasonsProps {
    items: Skipped[]
}

const SkippedReasons: FC<SkippedReasonsProps> = props => {
    const { items } = props

    return (
        <>
            {items.length > 0 && <H3 className="mx-3">Some results skipped:</H3>}
            {items.map((skipped, index) => (
                <SkippedMessage
                    key={skipped.reason}
                    skipped={skipped}
                    // Start with first item open, but only if it's not info severity or if there's only one item
                    startOpen={index === 0 && (skipped.severity !== 'info' || items.length === 1)}
                />
            ))}
        </>
    )
}

const SlimSkippedReasons: FC<SkippedReasonsProps> = props => {
    const { items } = props

    return (
        <div className={styles.streamingSkippedItem}>
            <Collapse openByDefault={items.length === 1}>
                {({ isOpen }) => (
                    <>
                        <CollapseHeader
                            as={Button}
                            outline={true}
                            variant="primary"
                            className={classNames(styles.button, 'p-2 w-100 bg-transparent border-0')}
                        >
                            <H4 className="d-flex align-items-center mb-0 w-100">
                                <Icon
                                    aria-label="Information"
                                    svgPath={mdiInformationOutline}
                                    className={classNames(styles.icon, 'flex-shrink-0')}
                                />

                                <span className="flex-grow-1 text-left">Why was the limit reached?</span>

                                <Icon
                                    aria-hidden={true}
                                    className={classNames('flex-shrink-0', styles.chevron)}
                                    svgPath={isOpen ? mdiChevronDown : mdiChevronLeft}
                                />
                            </H4>
                        </CollapseHeader>

                        <CollapsePanel className={styles.nestedCollapseReasons}>
                            {items.map((skipped, index) => (
                                <SkippedMessage
                                    key={skipped.reason}
                                    skipped={skipped}
                                    // Start with first item open, but only if it's not info severity or if there's only one item
                                    startOpen={index === 0 && (skipped.severity !== 'info' || items.length === 1)}
                                />
                            ))}
                        </CollapsePanel>

                        {!isOpen && <div className={classNames(styles.bottomBorderSpacer, 'mt-1')} />}
                    </>
                )}
            </Collapse>
        </div>
    )
}

interface SkippedItemsSearchProps {
    slim: boolean
    items: Skipped[]
    disabled: boolean
    onSearchSettingsChange: (event: React.FormEvent<HTMLInputElement>) => void
    onSubmit: (event: React.FormEvent) => void
}

const SkippedItemsSearch: FC<SkippedItemsSearchProps> = props => {
    const { slim, items, disabled, onSearchSettingsChange, onSubmit } = props

    return (
        <Form className="pb-3 px-3" onSubmit={onSubmit} data-testid="popover-form">
            <div className="mb-2 mt-3">Search again:</div>
            <div className="form-check">
                {items.map(
                    (skipped, index) =>
                        skipped.suggested && (
                            <Checkbox
                                key={skipped.suggested.queryExpression}
                                value={skipped.suggested.queryExpression}
                                onChange={onSearchSettingsChange}
                                data-testid={`streaming-progress-skipped-suggest-check-${index}`}
                                id={`streaming-progress-skipped-suggest-check-${index}`}
                                wrapperClassName="mb-1 d-block"
                                label={
                                    <>
                                        {skipped.suggested.title} (
                                        <SyntaxHighlightedSearchQuery query={skipped.suggested.queryExpression} />)
                                    </>
                                }
                            />
                        )
                )}
            </div>

            <Button
                type="submit"
                className="mt-2"
                variant={slim ? 'secondary' : 'primary'}
                disabled={disabled}
                size={slim ? 'sm' : undefined}
                data-testid="skipped-popover-form-submit-btn"
            >
                <Icon aria-hidden={true} className="mr-1" svgPath={mdiMagnify} />
                {slim ? <>Modify and re-run</> : <>Search again</>}
            </Button>
        </Form>
    )
}

interface ExhaustiveSearchMessageProps {
    query: string
}

export const ExhaustiveSearchMessage: FC<ExhaustiveSearchMessageProps> = props => {
    const { query } = props

    const validationErrors = useMemo(() => validateQueryForExhaustiveSearch(query), [query])

    return (
        <section className={styles.exhaustiveSearch}>
            <header className={styles.exhaustiveSearchHeader}>
                <Text className="m-0">Create a search job:</Text>
                <ProductStatusBadge status="experimental" />
            </header>

            {validationErrors.length > 0 && (
                <Alert variant="warning">
                    <ul className={styles.exhaustiveSearchWarningList}>
                        {validationErrors.map(validationError => (
                            <li key={validationError.reason}>{validationError.reason}</li>
                        ))}
                    </ul>
                </Alert>
            )}

            <Text className={classNames(validationErrors.length > 0 && 'text-muted', styles.exhaustiveSearchText)}>
                Search jobs exhaustively return all matches of a query. Results can be downloaded via CSV.
            </Text>
            <Button variant="secondary" size="sm" disabled={validationErrors.length > 0} className="mt-2">
                <Icon aria-hidden={true} className="mr-1" svgPath={mdiMagnify} />
                Create a search job
            </Button>
        </section>
    )
}
