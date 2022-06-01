import React from 'react'

import { Typography } from '@sourcegraph/wildcard'

import { InputTooltip } from '../../../../components/InputTooltip'

export interface BatchChangeChangesetsHeaderProps {
    allSelected?: boolean
    toggleSelectAll?: () => void
    disabled?: boolean
}

export const BatchChangeChangesetsHeader: React.FunctionComponent<
    React.PropsWithChildren<BatchChangeChangesetsHeaderProps>
> = ({ allSelected, toggleSelectAll, disabled }) => (
    <>
        <span className="d-none d-md-block" />
        {toggleSelectAll && (
            <InputTooltip
                type="checkbox"
                className="ml-2"
                checked={allSelected}
                onChange={toggleSelectAll}
                disabled={!!disabled}
                tooltip={
                    disabled ? 'You do not have permission to perform this operation' : 'Click to select all changesets'
                }
                aria-label={
                    disabled ? 'You do not have permission to perform this operation' : 'Click to select all changesets'
                }
            />
        )}
        <Typography.H5 as={Typography.H3} className="p-2 d-none d-md-block text-uppercase text-center text-nowrap">
            Status
        </Typography.H5>
        <Typography.H5 as={Typography.H3} className="p-2 d-none d-md-block text-uppercase text-nowrap">
            Changeset information
        </Typography.H5>
        <Typography.H5 as={Typography.H3} className="p-2 d-none d-md-block text-uppercase text-center text-nowrap">
            Check state
        </Typography.H5>
        <Typography.H5 as={Typography.H3} className="p-2 d-none d-md-block text-uppercase text-center text-nowrap">
            Review state
        </Typography.H5>
        <Typography.H5 as={Typography.H3} className="p-2 d-none d-md-block text-uppercase text-center text-nowrap">
            Changes
        </Typography.H5>
    </>
)
