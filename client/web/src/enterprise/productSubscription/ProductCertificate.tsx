import classNames from 'classnames'
import React from 'react'

import styles from './ProductCertificate.module.scss'

interface Props {
    /** The title of the certificate. */
    title: React.ReactFragment

    /** The subtitle of the certificate. */
    subtitle?: React.ReactFragment | null

    /** The detail text of the certificate. */
    detail?: React.ReactFragment | null

    /** Rendered after the certificate body (usually consists of a Bootstrap .card-footer). */
    footer?: React.ReactFragment | null

    className?: string
}

/**
 * Displays an official-looking certificate (with a Sourcegraph logo "watermark" background) with information about
 * the product license or subscription.
 *
 * In most cases, you should use a component that wraps this component and handles fetching the data to display.
 * Such components exist; check this component's TypeScript references.
 */
export const ProductCertificate: React.FunctionComponent<Props> = ({
    title,
    subtitle,
    detail,
    footer,
    className = '',
}) => (
    <div className={classNames('card', className, 'test-product-certificate')}>
        <div className="card-body d-flex align-items-center">
            <img
                className={classNames(styles.logo, 'mr-1', 'p-2')}
                src="/.assets/img/sourcegraph-mark.svg?v2"
                alt="Sourcegraph logo"
            />
            <div>
                <h2 className="font-weight-normal mb-1">{title}</h2>
                {subtitle && <h3 className="text-muted font-weight-normal">{subtitle}</h3>}
                {detail && <p className="text-muted mb-0">{detail}</p>}
            </div>
        </div>
        <div className={styles.bg} />
        {footer && <div className={styles.footer}>{footer}</div>}
    </div>
)
