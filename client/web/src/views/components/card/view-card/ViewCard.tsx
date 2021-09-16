import classnames from 'classnames'
import React, { PropsWithChildren, ReactNode } from 'react'
import { useLocation } from 'react-router-dom'

import { ViewProviderResult } from '@sourcegraph/shared/src/api/extension/extensionHostApi'
import { isErrorLike } from '@sourcegraph/shared/src/util/errors'

import { ErrorBoundary } from '../../../../components/ErrorBoundary'

import styles from './ViewCard.module.scss'

type ViewCardElementProps = React.DetailedHTMLProps<Omit<React.HTMLAttributes<HTMLElement>, 'contextMenu'>, HTMLElement>

export interface ViewCardProps extends ViewCardElementProps {
    /**
     * Insight data (title, chart content)
     */
    insight: ViewProviderResult

    /**
     * Custom card actions (like filter buttons) that render element right next to three dots
     * card context menu.
     */
    actions?: ReactNode

    /**
     * Ref prop for root element (section) of insight content card.
     */
    innerRef?: React.RefObject<HTMLElement>

    /**
     * Custom c
     */
    contextMenu?: ReactNode
}

/**
 * Renders insight card content. Loading state, error state and insight itself.
 */
export const ViewCard: React.FunctionComponent<PropsWithChildren<ViewCardProps>> = props => {
    const {
        insight: { id, view },
        actions,
        contextMenu,
        children,
        innerRef,
        ...otherProps
    } = props

    const location = useLocation()
    const title = !isErrorLike(view) ? view?.title : null
    const subtitle = !isErrorLike(view) ? view?.subtitle : null

    // In case if we don't have a content for the header component
    // we should render nothing
    const hasHeader = title || subtitle || contextMenu || actions

    return (
        <section
            {...otherProps}
            data-testid={`insight-card.${id}`}
            /* eslint-disable-next-line jsx-a11y/no-noninteractive-tabindex */
            tabIndex={0}
            ref={innerRef}
            className={classnames('card', otherProps.className, styles.viewCard)}
        >
            <ErrorBoundary
                className="pt-0"
                location={location}
                extraContext={
                    <>
                        <p>ID: {id}</p>
                        <pre>View: {JSON.stringify(view, null, 2)}</pre>
                    </>
                }
            >
                {hasHeader && (
                    <header className={styles.viewCardHeader}>
                        <div className={styles.viewCardHeaderContent}>
                            <h4 className={styles.viewCardTitle}>{title}</h4>
                            {subtitle && <div className={styles.viewCardSubtitle}>{subtitle}</div>}
                        </div>

                        <div className="align-self-start d-flex align-items-center">
                            {actions}
                            {contextMenu}
                        </div>
                    </header>
                )}

                {children}
            </ErrorBoundary>
        </section>
    )
}
