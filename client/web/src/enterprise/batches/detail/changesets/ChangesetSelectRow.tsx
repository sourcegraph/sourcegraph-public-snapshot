import InfoCircleOutlineIcon from 'mdi-react/InfoCircleOutlineIcon'
import React from 'react'

import { isErrorLike } from '@sourcegraph/shared/src/util/errors'
import { pluralize } from '@sourcegraph/shared/src/util/strings'

import { ErrorAlert } from '../../../../components/alerts'

export interface ChangesetSelectRowProps {
    selected: Set<string>
    onSubmit: () => void
    isSubmitting: boolean | Error
}

export const ChangesetSelectRow: React.FunctionComponent<ChangesetSelectRowProps> = ({
    selected,
    onSubmit,
    isSubmitting,
}) => (
    <>
        <div className="row align-items-center no-gutters">
            <div className="ml-2 col">
                <InfoCircleOutlineIcon className="icon-inline text-muted mr-2" />
                Select changesets to detach them
            </div>
            <div className="w-100 d-block d-md-none" />
            <div className="m-0 col col-md-auto">
                <div className="row no-gutters">
                    <div className="col my-2 ml-0 ml-sm-2">
                        <button
                            type="button"
                            className="btn btn-secondary text-nowrap"
                            onClick={onSubmit}
                            disabled={selected.size === 0 || isSubmitting === true}
                        >
                            {selected.size > 0
                                ? `Detach ${selected.size} ${pluralize('changeset', selected.size)}`
                                : 'Detach changesets'}
                        </button>
                    </div>
                </div>
            </div>
        </div>

        {isErrorLike(isSubmitting) && <ErrorAlert error={isSubmitting} />}
    </>
)
