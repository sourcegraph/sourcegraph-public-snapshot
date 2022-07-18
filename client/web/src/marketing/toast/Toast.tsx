import * as React from 'react'

import { mdiClose } from '@mdi/js'
import classNames from 'classnames'

import { Button, CardTitle, CardBody, Card, Icon, H3 } from '@sourcegraph/wildcard'

import styles from './Toast.module.scss'

interface ToastProps {
    title?: React.ReactNode
    subtitle?: React.ReactNode
    cta?: JSX.Element
    footer?: JSX.Element
    onDismiss: () => void
    className?: string
    toastBodyClassName?: string
    toastContentClassName?: string
}

export const Toast: React.FunctionComponent<React.PropsWithChildren<ToastProps>> = props => (
    <Card className={classNames(styles.toast, props.className)}>
        <CardBody className={classNames(styles.toastBody, props.toastBodyClassName)}>
            <div className={classNames('d-flex justify-content-end', styles.closeButtonWrap)}>
                <Button onClick={props.onDismiss} variant="icon" className="test-close-toast" aria-label="Close">
                    <Icon className={styles.closeButtonIcon} aria-hidden={true} svgPath={mdiClose} />
                </Button>
            </div>
            {props.title && (
                <CardTitle as="header" className="d-flex align-items-center mb-1">
                    <H3 className="mb-0">{props.title}</H3>
                </CardTitle>
            )}
            {props.subtitle}
            {props.cta && (
                <div className={classNames(styles.contentsCta, props.toastContentClassName)}>{props.cta}</div>
            )}
        </CardBody>
        {props.footer && <div className={classNames(styles.footer)}>{props.footer}</div>}
    </Card>
)
