import type { FunctionComponent, HTMLAttributes, PropsWithChildren, ReactElement } from 'react'

import { mdiChevronDown, mdiChevronUp } from '@mdi/js'
import classNames from 'classnames'

import { SyntaxHighlightedSearchQuery } from '@sourcegraph/branded'
import { Button, Collapse, CollapseHeader, CollapsePanel, Icon } from '@sourcegraph/wildcard'

import { TruncatedText } from '../../../../../../trancated-text/TruncatedText'

import styles from './FilterCollapseSection.module.scss'

interface FilterCollapseSectionProps extends HTMLAttributes<HTMLButtonElement> {
    open: boolean
    title: string
    preview: string
    hasActiveFilter: boolean
    className?: string
    withSeparators: boolean
    onOpenChange: (opened: boolean) => void
}

export function FilterCollapseSection(props: PropsWithChildren<FilterCollapseSectionProps>): ReactElement {
    const { open, title, preview, hasActiveFilter, className, withSeparators, children, onOpenChange, ...attributes } =
        props

    return (
        <div className={classNames(className, { [styles.rootNoCollapse]: !withSeparators })}>
            <Collapse isOpen={open} onOpenChange={onOpenChange}>
                <CollapseHeader {...attributes} as={Button} outline={true} className={styles.collapseButton}>
                    <Icon
                        aria-hidden={true}
                        className={styles.collapseIcon}
                        svgPath={open ? mdiChevronUp : mdiChevronDown}
                    />

                    <span className={styles.buttonText}>{title}</span>

                    {!open && preview && <FilterPreviewPill text={preview} className={styles.panelPreview} />}
                    {hasActiveFilter && <div className={styles.changedFilterMarker} />}
                </CollapseHeader>

                {open && <CollapsePanel className={styles.collapsePanel}>{children}</CollapsePanel>}

                {withSeparators && <hr />}
            </Collapse>
        </div>
    )
}

export interface FilterPreviewPillProps {
    text: string
    className?: string
}

export const FilterPreviewPill: FunctionComponent<FilterPreviewPillProps> = props => {
    const { text, className } = props

    return (
        <TruncatedText className={classNames(className, styles.filterBadge)}>
            <SyntaxHighlightedSearchQuery query={text} />
        </TruncatedText>
    )
}
