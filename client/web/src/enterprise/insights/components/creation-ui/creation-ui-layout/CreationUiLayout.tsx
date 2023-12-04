import { forwardRef } from 'react'

import classNames from 'classnames'

import { Container, type ForwardReferenceComponent } from '@sourcegraph/wildcard'

import styles from './CreationUiLayout.module.scss'

export const CreationUiLayout = forwardRef((props, reference) => {
    const { as: Component = 'div', className, ...attributes } = props

    return <Component ref={reference} {...attributes} className={classNames(styles.root, className)} />
}) as ForwardReferenceComponent<'div', {}>

export const CreationUIForm = forwardRef((props, reference) => {
    const { as: Component = 'form', ...attributes } = props

    return (
        <Container className={styles.rootForm}>
            <Component ref={reference} {...attributes} />
        </Container>
    )
}) as ForwardReferenceComponent<'form', {}>

export const CreationUIPreview = forwardRef((props, reference) => {
    const { as: Component = 'aside', className, ...attributes } = props

    return (
        <Component
            {...attributes}
            ref={reference}
            aria-label="Code Insight live preview"
            className={classNames(styles.rootLivePreview, className)}
        />
    )
}) as ForwardReferenceComponent<'aside', {}>
