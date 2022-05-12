import React from 'react'

import classNames from 'classnames'

import { CardBody, Card, Typography } from '@sourcegraph/wildcard'

import styles from './ProductCertificate.module.scss'

interface Props {
    /** The title of the certificate. */
    title: React.ReactFragment

    /** The subtitle of the certificate. */
    subtitle?: React.ReactFragment | null

    /** The detail text of the certificate. */
    detail?: React.ReactFragment | null

    /** Rendered after the certificate body (usually consists of a Wildcard <CardFooter />). */
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
export const ProductCertificate: React.FunctionComponent<React.PropsWithChildren<Props>> = ({
    title,
    subtitle,
    detail,
    footer,
    className = '',
}) => (
    <Card className={className} data-testid="product-certificate">
        <CardBody className="d-flex align-items-center">
            <img
                className={classNames(styles.logo, 'mr-1', 'p-2')}
                src="/.assets/img/sourcegraph-mark.svg?v2"
                alt="Sourcegraph logo"
            />
            <div>
                <Typography.H2 className="font-weight-normal mb-1">{title}</Typography.H2>
                {subtitle && <Typography.H3 className="text-muted font-weight-normal">{subtitle}</Typography.H3>}
                {detail && <p className="text-muted mb-0">{detail}</p>}
            </div>
        </CardBody>
        <div className={styles.bg} />
        {footer && <div className={styles.footer}>{footer}</div>}
    </Card>
)
