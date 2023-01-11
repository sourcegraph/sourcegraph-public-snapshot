import React from 'react'

import { H3, H5 } from '@sourcegraph/wildcard'

import { InputTooltip } from '../../../../components/InputTooltip'

import styles from './PreviewListHeader.module.scss'

export interface PreviewListHeaderProps {
    allSelected?: boolean
    toggleSelectAll?: () => void
}

export const PreviewListHeader: React.FunctionComponent<React.PropsWithChildren<PreviewListHeaderProps>> = ({
    allSelected,
    toggleSelectAll,
}) => (
    <li className={styles.listItem}>
        <span className="p-2 d-none d-sm-block" />
        {toggleSelectAll && (
            <div className="d-flex p-2 align-items-center">
                {/* eslint-disable-next-line no-restricted-syntax*/}
                <InputTooltip
                    type="checkbox"
                    checked={allSelected}
                    onChange={toggleSelectAll}
                    tooltip="Click to select all changesets"
                    aria-label="Click to select all changesets"
                    placement="right"
                />
                <span className="pl-2 d-block d-sm-none">Select all</span>
            </div>
        )}
        <H5 as={H3} className="p-2 d-none d-sm-block text-uppercase text-center" aria-hidden={true}>
            Current state
        </H5>
        <H5 as={H3} className="d-none d-sm-block text-uppercase text-center" aria-hidden={true}>
            +<br />-
        </H5>
        <H5 as={H3} className="p-2 d-none d-sm-block text-uppercase text-nowrap" aria-hidden={true}>
            Actions
        </H5>
        <H5 as={H3} className="p-2 d-none d-sm-block text-uppercase text-nowrap" aria-hidden={true}>
            Changeset information
        </H5>
        <H5 as={H3} className="p-2 d-none d-sm-block text-uppercase text-center text-nowrap" aria-hidden={true}>
            Commit changes
        </H5>
        <H5 as={H3} className="p-2 d-none d-sm-block text-uppercase text-center text-nowrap" aria-hidden={true}>
            Change state
        </H5>
    </li>
)
