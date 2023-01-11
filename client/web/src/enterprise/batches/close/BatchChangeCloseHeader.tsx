import React from 'react'

import classNames from 'classnames'

import { H2, H3, H5 } from '@sourcegraph/wildcard'

import styles from './BatchChangeCloseHeader.module.scss'

export interface BatchChangeCloseHeaderProps {
    // Nothing.
}

const BatchChangeCloseHeader: React.FunctionComponent<React.PropsWithChildren<BatchChangeCloseHeaderProps>> = () => (
    <>
        <span className="d-none d-md-block" />
        <H5 as={H3} aria-hidden={true} className="d-none d-md-block text-uppercase text-center text-nowrap">
            Action
        </H5>
        <H5 as={H3} aria-hidden={true} className="d-none d-md-block text-uppercase text-nowrap">
            Changeset information
        </H5>
        <H5 as={H3} aria-hidden={true} className="d-none d-md-block text-uppercase text-center text-nowrap">
            Check state
        </H5>
        <H5 as={H3} aria-hidden={true} className="d-none d-md-block text-uppercase text-center text-nowrap">
            Review state
        </H5>
        <H5 as={H3} aria-hidden={true} className="d-none d-md-block text-uppercase text-center text-nowrap">
            Changes
        </H5>
    </>
)

export const BatchChangeCloseHeaderWillCloseChangesets: React.FunctionComponent<
    React.PropsWithChildren<BatchChangeCloseHeaderProps>
> = () => (
    <>
        <H2 className={classNames(styles.batchChangeCloseHeaderRow, 'test-batches-close-willclose-header')}>
            Closing the batch change will close the following changesets:
        </H2>
        <BatchChangeCloseHeader />
    </>
)

export const BatchChangeCloseHeaderWillKeepChangesets: React.FunctionComponent<
    React.PropsWithChildren<BatchChangeCloseHeaderProps>
> = () => (
    <>
        <H2 className={styles.batchChangeCloseHeaderRow}>The following changesets will remain open:</H2>
        <BatchChangeCloseHeader />
    </>
)
