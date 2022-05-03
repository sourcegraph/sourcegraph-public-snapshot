import { PropsWithChildren, ReactElement } from 'react'

import ChevronDownIcon from 'mdi-react/ChevronDownIcon'
import ChevronUpIcon from 'mdi-react/ChevronUpIcon'

import { SyntaxHighlightedSearchQuery } from '@sourcegraph/search-ui'
import { Button, Collapse, CollapseHeader, CollapsePanel, Icon } from '@sourcegraph/wildcard'

import { TruncatedText } from '../../../../../../trancated-text/TruncatedText'

import styles from './FilterCollapseSection.module.scss'

interface FilterCollapseSectionProps {
    open: boolean
    title: string
    preview: string
    hasActiveFilter: boolean
    className?: string
    onOpenChange: (opened: boolean) => void
}

export function FilterCollapseSection(props: PropsWithChildren<FilterCollapseSectionProps>): ReactElement {
    const { open, title, preview, hasActiveFilter, className, children, onOpenChange } = props

    return (
        <Collapse isOpen={open} onOpenChange={onOpenChange}>
            <div className={className}>
                <CollapseHeader
                    as={Button}
                    aria-label={open ? 'Expand' : 'Collapse'}
                    outline={true}
                    className={styles.collapseButton}
                >
                    <Icon className={styles.collapseIcon} as={open ? ChevronUpIcon : ChevronDownIcon} />

                    <span className={styles.buttonText}>{title}</span>

                    {!open && preview && (
                        <TruncatedText className={styles.filterBadge}>
                            <SyntaxHighlightedSearchQuery query={preview} />
                        </TruncatedText>
                    )}
                    {hasActiveFilter && <div className={styles.changedFilterMarker} />}
                </CollapseHeader>

                {open && <CollapsePanel className={styles.collapsePanel}>{children}</CollapsePanel>}
            </div>

            <hr />
        </Collapse>
    )
}
