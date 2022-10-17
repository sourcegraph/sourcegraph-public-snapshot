import React from 'react'

import { H3, H5 } from '@sourcegraph/wildcard'

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
        <H5 as={H3} className="p-2 d-none d-sm-block text-uppercase text-center">
            Current state
        </H5>
        <H5 as={H3} className="d-none d-sm-block text-uppercase text-center">
            +<br />-
        </H5>
        <H5 as={H3} className="p-2 d-none d-sm-block text-uppercase text-nowrap">
            Actions
        </H5>
        <H5 as={H3} className="p-2 d-none d-sm-block text-uppercase text-nowrap">
            Changeset information
        </H5>
        <H5 as={H3} className="p-2 d-none d-sm-block text-uppercase text-center text-nowrap">
            Commit changes
        </H5>
        <H5 as={H3} className="p-2 d-none d-sm-block text-uppercase text-center text-nowrap">
            Change state
        </H5>
    </>
)
