import classNames from 'classnames'
import React, { PropsWithChildren, ReactNode } from 'react'
import { useLocation } from 'react-router-dom'

import { ErrorBoundary } from '../../../components/ErrorBoundary'

import styles from './View.module.scss'

const stopPropagation = (event: React.MouseEvent): void => event.stopPropagation()

type ViewCardElementProps = React.DetailedHTMLProps<Omit<React.HTMLAttributes<HTMLElement>, 'contextMenu'>, HTMLElement>

export interface ViewCardProps extends ViewCardElementProps {
    title?: string
    subtitle?: string
    innerRef?: React.RefObject<HTMLElement>

    /**
     * Custom card actions (like filter buttons) that render element right next to three dots
     * card context menu.
     */
    actions?: ReactNode
}

export const View: React.FunctionComponent<PropsWithChildren<ViewCardProps>> = props => {
    const { title, subtitle, actions, children, innerRef, ...otherProps } = props

    // In case if we don't have a content for the header component
    // we should render nothing
    const hasHeader = title || subtitle || actions

    return (
        <section
            {...otherProps}
            /* eslint-disable-next-line jsx-a11y/no-noninteractive-tabindex */
            tabIndex={0}
            ref={innerRef}
            className={classNames('card', otherProps.className, styles.view)}
        >
            <ErrorBoundary className="pt-0" location={useLocation()}>
                {hasHeader && (
                    <header className={styles.header}>
                        <div className={styles.headerContent}>
                            <h4 title={title} className={styles.title}>
                                {title}
                            </h4>
                            {subtitle}
                        </div>

                        {/* eslint-disable-next-line jsx-a11y/click-events-have-key-events,jsx-a11y/no-static-element-interactions */}
                        <div
                            // View component usually is rendered within a view grid component. To suppress
                            // bad click that lead to card DnD events in view grid we stop event bubbling for
                            // clicks.
                            onClick={stopPropagation}
                            onMouseDown={stopPropagation}
                            className={styles.action}
                        >
                            {actions}
                        </div>
                    </header>
                )}

                {children}
            </ErrorBoundary>
        </section>
    )
}
