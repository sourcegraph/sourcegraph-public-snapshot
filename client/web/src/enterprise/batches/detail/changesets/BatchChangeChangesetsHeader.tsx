import React from 'react'

import { H3, H5 } from '@sourcegraph/wildcard'

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
            // eslint-disable-next-line no-restricted-syntax
            <InputTooltip
                type="checkbox"
                className="ml-2"
                checked={allSelected}
                onChange={toggleSelectAll}
                disabled={!!disabled}
                placement="right"
                tooltip={
                    disabled ? 'You do not have permission to perform this operation' : 'Click to select all changesets'
                }
                aria-label={
                    disabled ? 'You do not have permission to perform this operation' : 'Click to select all changesets'
                }
            />
        )}
        <H5 as={H3} className="p-2 d-none d-md-block text-uppercase text-center text-nowrap">
            Status
        </H5>
        <H5 as={H3} className="p-2 d-none d-md-block text-uppercase text-nowrap">
            Changeset information
        </H5>
        <H5 as={H3} className="p-2 d-none d-md-block text-uppercase text-center text-nowrap">
            Check state
        </H5>
        <H5 as={H3} className="p-2 d-none d-md-block text-uppercase text-center text-nowrap">
            Review state
        </H5>
        <H5 as={H3} className="p-2 d-none d-md-block text-uppercase text-center text-nowrap">
            Changes
        </H5>
    </>
)
