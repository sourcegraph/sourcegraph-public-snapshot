import { isErrorLike } from '@sourcegraph/shared/src/util/errors'
import { pluralize } from '@sourcegraph/shared/src/util/strings'
import React from 'react'
import { ErrorAlert } from '../../../../components/alerts'

export interface ChangesetSelectRowProps {
    selected: Set<string>
    onSubmit: () => void
    deselectAll: () => void
    isSubmitting: boolean | Error
}

export const ChangesetSelectRow: React.FunctionComponent<ChangesetSelectRowProps> = ({
    selected,
    onSubmit,
    deselectAll,
    isSubmitting,
}) => (
    <>
        <div className="row align-items-center no-gutters">
            <div className="ml-2 col-auto">
                <input
                    id="deselect-all"
                    type="checkbox"
                    className="btn"
                    checked={selected.size > 0}
                    onChange={deselectAll}
                    data-tooltip="Click to deselect all"
                />
            </div>
            <div className="ml-2 col">
                {selected.size} archived {pluralize('changeset', selected.size)} selected for detaching.
            </div>
            <div className="w-100 d-block d-md-none" />
            <div className="m-0 col col-md-auto">
                <div className="row no-gutters">
                    <div className="col my-2 ml-0 ml-sm-2">
                        <button
                            type="button"
                            className="btn btn-primary text-nowrap"
                            onClick={onSubmit}
                            disabled={isSubmitting === true}
                            data-tooltip={`Click to detach ${selected.size} ${pluralize('changeset', selected.size)}.`}
                        >
                            Detach {selected.size} {pluralize('changeset', selected.size)}
                        </button>
                    </div>
                </div>
            </div>
        </div>

        {isErrorLike(isSubmitting) && <ErrorAlert error={isSubmitting} />}
    </>
)
