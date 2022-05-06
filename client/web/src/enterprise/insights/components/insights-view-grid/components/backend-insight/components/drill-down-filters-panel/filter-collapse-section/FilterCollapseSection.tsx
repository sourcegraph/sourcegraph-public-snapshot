import { PropsWithChildren, ReactElement } from 'react'

import classNames from 'classnames'
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
    withSeparators: boolean
    onOpenChange: (opened: boolean) => void
}

export function FilterCollapseSection(props: PropsWithChildren<FilterCollapseSectionProps>): ReactElement {
    const { open, title, preview, hasActiveFilter, className, withSeparators, children, onOpenChange } = props

    return (
        <div className={classNames(className, { [styles.rootNoCollapse]: !withSeparators })}>
            <Collapse isOpen={open} onOpenChange={onOpenChange}>
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

                {withSeparators && <hr />}
            </Collapse>
        </div>
    )
}
