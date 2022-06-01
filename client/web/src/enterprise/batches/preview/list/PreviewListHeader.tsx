import React from 'react'

import { Typography } from '@sourcegraph/wildcard'

import { InputTooltip } from '../../../../components/InputTooltip'

export interface PreviewListHeaderProps {
    allSelected?: boolean
    toggleSelectAll?: () => void
}

export const PreviewListHeader: React.FunctionComponent<React.PropsWithChildren<PreviewListHeaderProps>> = ({
    allSelected,
    toggleSelectAll,
}) => (
    <>
        <span className="p-2 d-none d-sm-block" />
        {toggleSelectAll && (
            <div className="d-flex p-2 align-items-center">
                <InputTooltip
                    type="checkbox"
                    checked={allSelected}
                    onChange={toggleSelectAll}
                    tooltip="Click to select all changesets"
                    aria-label="Click to select all changesets"
                />
                <span className="pl-2 d-block d-sm-none">Select all</span>
            </div>
        )}
        <Typography.H5 as={Typography.H3} className="p-2 d-none d-sm-block text-uppercase text-center">
            Current state
        </Typography.H5>
        <Typography.H5 as={Typography.H3} className="d-none d-sm-block text-uppercase text-center">
            +<br />-
        </Typography.H5>
        <Typography.H5 as={Typography.H3} className="p-2 d-none d-sm-block text-uppercase text-nowrap">
            Actions
        </Typography.H5>
        <Typography.H5 as={Typography.H3} className="p-2 d-none d-sm-block text-uppercase text-nowrap">
            Changeset information
        </Typography.H5>
        <Typography.H5 as={Typography.H3} className="p-2 d-none d-sm-block text-uppercase text-center text-nowrap">
            Commit changes
        </Typography.H5>
        <Typography.H5 as={Typography.H3} className="p-2 d-none d-sm-block text-uppercase text-center text-nowrap">
            Change state
        </Typography.H5>
    </>
)
