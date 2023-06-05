import React from 'react'

import { mdiChevronLeft, mdiChevronRight } from '@mdi/js'
import classNames from 'classnames'

import { Button, Icon, Tooltip } from '@sourcegraph/wildcard'

import { NavItem } from '../nav'

import { HistoryStack } from './useHistoryStack'

import navBarStyles from '../nav/NavBar/NavBar.module.scss'
import navItemStyles from '../nav/NavBar/NavItem.module.scss'
import styles from './TauriNavigation.module.scss'

export const TauriNavigation: React.FC<{ historyStack: HistoryStack }> = ({ historyStack }) => (
    <>
        <ul className={classNames(navBarStyles.list, styles.list)}>
            <NavItem>
                <Tooltip content={historyStack.canGoBack && 'Go back'}>
                    <Button
                        variant="icon"
                        className={classNames(styles.button, navItemStyles.link)}
                        disabled={!historyStack.canGoBack}
                        onClick={historyStack.goBack}
                    >
                        <span className={navItemStyles.linkContent}>
                            <Icon svgPath={mdiChevronLeft} aria-label="Back" className={navItemStyles.icon} />
                        </span>
                    </Button>
                </Tooltip>
            </NavItem>
            <NavItem>
                <Tooltip content={historyStack.canGoForward && 'Go forward'}>
                    <Button
                        variant="icon"
                        className={classNames(styles.button, navItemStyles.link)}
                        onClick={historyStack.goForward}
                        disabled={!historyStack.canGoForward}
                    >
                        <span className={navItemStyles.linkContent}>
                            <Icon svgPath={mdiChevronRight} aria-label="Forward" className={navItemStyles.icon} />
                        </span>
                    </Button>
                </Tooltip>
            </NavItem>
        </ul>
        <hr className={navBarStyles.divider} aria-hidden={true} />
    </>
)
