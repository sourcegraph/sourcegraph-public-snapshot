import React from 'react'

import classNames from 'classnames'

import { Typography } from '@sourcegraph/wildcard'

import styles from './BatchChangeCloseHeader.module.scss'

export interface BatchChangeCloseHeaderProps {
    // Nothing.
}

const BatchChangeCloseHeader: React.FunctionComponent<React.PropsWithChildren<BatchChangeCloseHeaderProps>> = () => (
    <>
        <span className="d-none d-md-block" />
        <Typography.H5 className="d-none d-md-block text-uppercase text-center text-nowrap">Action</Typography.H5>
        <Typography.H5 className="d-none d-md-block text-uppercase text-nowrap">Changeset information</Typography.H5>
        <Typography.H5 className="d-none d-md-block text-uppercase text-center text-nowrap">Check state</Typography.H5>
        <Typography.H5 className="d-none d-md-block text-uppercase text-center text-nowrap">Review state</Typography.H5>
        <Typography.H5 className="d-none d-md-block text-uppercase text-center text-nowrap">Changes</Typography.H5>
    </>
)

export const BatchChangeCloseHeaderWillCloseChangesets: React.FunctionComponent<
    React.PropsWithChildren<BatchChangeCloseHeaderProps>
> = () => (
    <>
        <Typography.H2 className={classNames(styles.batchChangeCloseHeaderRow, 'test-batches-close-willclose-header')}>
            Closing the batch change will close the following changesets:
        </Typography.H2>
        <BatchChangeCloseHeader />
    </>
)

export const BatchChangeCloseHeaderWillKeepChangesets: React.FunctionComponent<
    React.PropsWithChildren<BatchChangeCloseHeaderProps>
> = () => (
    <>
        <Typography.H2 className={styles.batchChangeCloseHeaderRow}>
            The following changesets will remain open:
        </Typography.H2>
        <BatchChangeCloseHeader />
    </>
)
