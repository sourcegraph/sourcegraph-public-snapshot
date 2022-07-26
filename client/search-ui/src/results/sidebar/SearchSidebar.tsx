import { FC, HTMLAttributes, PropsWithChildren } from 'react'

import { mdiClose } from '@mdi/js'
import classNames from 'classnames'
import StickyBox from 'react-sticky-box'

import { Button, Icon } from '@sourcegraph/wildcard'

import styles from './SearchSidebar.module.scss'

interface SearchSidebarProps extends HTMLAttributes<HTMLElement> {
    /** Call whenever the user clicks the close side panel button */
    onClose?: () => void
}

export const SearchSidebar: FC<PropsWithChildren<SearchSidebarProps>> = props => {
    const { children, className, onClose, ...attributes } = props

    return (
        <aside
            {...attributes}
            className={classNames(styles.sidebar, className)}
            role="region"
            aria-label="Search sidebar"
        >
            <StickyBox className={styles.stickyBox} offsetTop={8}>
                <div className={styles.header}>
                    <Button variant="icon" onClick={onClose}>
                        <Icon svgPath={mdiClose} aria-label="Close sidebar" />
                    </Button>
                </div>
                {children}
            </StickyBox>
        </aside>
    )
}
