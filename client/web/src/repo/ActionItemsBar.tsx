import React from 'react'

import { mdiMenuUp, mdiMenuDown } from '@mdi/js'
import classNames from 'classnames'
import * as H from 'history'

import { Button, Icon } from '@sourcegraph/wildcard'

import { ErrorBoundary } from '../components/ErrorBoundary'
import { useCarousel } from '../components/useCarousel'
import { GitBlameButton } from '../extensions/git/GitBlameButton'

import styles from './ActionItemsBar.module.scss'

interface ActionItemsBarProps {
    useActionItemsBar: () => { isOpen: boolean | undefined; barReference: React.RefCallback<HTMLElement> }
    location: H.Location
}

export const ActionItemsBar: React.FC<ActionItemsBarProps> = React.memo(props => {
    const { isOpen, barReference } = props.useActionItemsBar()

    const {
        carouselReference,
        canScrollNegative,
        canScrollPositive,
        onNegativeClicked,
        onPositiveClicked,
    } = useCarousel({ direction: 'topToBottom' })

    if (!isOpen) {
        return <div className={styles.barCollapsed} />
    }

    return (
        <div className={classNames('p-0 mr-2 position-relative d-flex flex-column', styles.bar)} ref={barReference}>
            {/* To be clear to users that this isn't an error reported by extensions about e.g. the code they're viewing. */}
            <ErrorBoundary location={props.location} render={error => <span>Component error: {error.message}</span>}>
                <ActionItemsDivider />
                {canScrollNegative && (
                    <Button
                        className={classNames('p-0 border-0', styles.scroll, styles.listItem)}
                        onClick={onNegativeClicked}
                        tabIndex={-1}
                        variant="link"
                        aria-label="Scroll up"
                    >
                        <Icon aria-hidden={true} svgPath={mdiMenuUp} />
                    </Button>
                )}
                <ul className={classNames('list-unstyled m-0', styles.list)} ref={carouselReference}>
                    <li className={styles.listItem}>
                        <GitBlameButton />
                    </li>
                </ul>
                {/* <ActionsContainer
                    menu={ContributableMenu.EditorTitle}
                    returnInactiveMenuItems={true}
                    extensionsController={props.extensionsController}
                    empty={null}
                    location={props.location}
                    platformContext={props.platformContext}rr
                    telemetryService={props.telemetryService}
                >
                    {items => (
                        <ul className={classNames('list-unstyled m-0', styles.list)} ref={carouselReference}>
                            {items.map((item, index) => {
                                const hasIconURL = !!item.action.actionItem?.iconURL
                                const className = classNames(
                                    actionItemClassName,
                                    !hasIconURL && classNames(styles.actionNoIcon, getIconClassName(index), 'text-sm')
                                )
                                const inactiveClassName = classNames(
                                    styles.actionInactive,
                                    !hasIconURL && styles.actionNoIconInactive
                                )
                                const listItemClassName = classNames(
                                    styles.listItem,
                                    index !== items.length - 1 && 'mb-1'
                                )

                                const dataContent = !hasIconURL ? item.action.category?.slice(0, 1) : undefined

                                return (
                                    <li key={item.action.id} className={listItemClassName}>
                                        <ActionItem
                                            {...props}
                                            {...item}
                                            className={className}
                                            dataContent={dataContent}
                                            variant="actionItem"
                                            iconClassName={styles.icon}
                                            pressedClassName={styles.actionPressed}
                                            inactiveClassName={inactiveClassName}
                                            hideLabel={true}
                                            tabIndex={-1}
                                            hideExternalLinkIcon={true}
                                            disabledDuringExecution={true}
                                        />
                                    </li>
                                )
                            })}
                        </ul>
                    )}
                </ActionsContainer> */}
                {canScrollPositive && (
                    <Button
                        className={classNames('p-0 border-0', styles.scroll, styles.listItem)}
                        onClick={onPositiveClicked}
                        tabIndex={-1}
                        variant="link"
                        aria-label="Scroll down"
                    >
                        <Icon aria-hidden={true} svgPath={mdiMenuDown} />
                    </Button>
                )}
            </ErrorBoundary>
        </div>
    )
})

ActionItemsBar.displayName = 'ActionItemsBar'

const ActionItemsDivider: React.FC<{ className?: string }> = ({ className }) => (
    <div className={classNames('position-relative rounded-sm d-flex', styles.dividerHorizontal, className)} />
)
