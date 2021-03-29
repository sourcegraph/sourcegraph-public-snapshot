import { pluralize } from '@sourcegraph/shared/src/util/strings'
import React from 'react'

export interface ChangesetSelectRowProps {
    selected: Set<string>
    onSubmit: () => void
    deselectAll: () => void
}

export const ChangesetSelectRow: React.FunctionComponent<ChangesetSelectRowProps> = ({
    selected,
    onSubmit,
    deselectAll,
}) => {
    return (
        <>
            <div className="row align-items-center no-gutters">
                <div className="ml-2 col-auto">
                    <input id={'deselect-all'} type="checkbox" className="btn" checked={true} onChange={deselectAll} />
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
                                className={'btn btn-primary text-nowrap'}
                                onClick={onSubmit}
                                data-tooltip={`Click to detach ${selected.size} ${pluralize(
                                    'changeset',
                                    selected.size
                                )}.`}
                            >
                                Detach {selected.size} {pluralize('changeset', selected.size)}
                            </button>
                        </div>
                    </div>
                </div>
            </div>
        </>
    )
}
