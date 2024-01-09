import { FC } from 'react'

import { mdiBook, mdiLink } from '@mdi/js'
import classNames from 'classnames'

import { Icon, Link, H4 } from '@sourcegraph/wildcard'

import styles from './FiltersDocFooter.module.scss'

interface FiltersDocFooterProps {
    className?: string
}

export const FiltersDocFooter: FC<FiltersDocFooterProps> = ({ className }) => (
    <footer className={classNames(styles.footer, className)}>
        <Link target="_blank" rel="noopener" to="/help/code_search/reference/queries" className={styles.link}>
            <Icon svgPath={mdiBook} aria-hidden={true} className={styles.linkIcon} />
            <span className={styles.linkText}>
                <H4 className="m-0">Need more advanced filters?</H4>
                <small>Explore the query syntax docs</small>
            </span>
            <Icon svgPath={mdiLink} aria-hidden={true} className={styles.linkIcon} />
        </Link>
    </footer>
)
