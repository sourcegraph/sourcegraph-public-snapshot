import { noop } from 'lodash'
import InfoCircleOutlineIcon from 'mdi-react/InfoCircleOutlineIcon'
import React, { useMemo, useContext } from 'react'

import { pluralize } from '@sourcegraph/shared/src/util/strings'
import { useObservable } from '@sourcegraph/shared/src/util/useObservable'
import { Button } from '@sourcegraph/wildcard'

import { BatchSpecApplyPreviewVariables, Scalars } from '../../../../graphql-operations'
import { Action, DropdownButton } from '../../DropdownButton'
import { MultiSelectContext } from '../../MultiSelectContext'
import { BatchChangePreviewContext } from '../BatchChangePreviewContext'

import { queryPublishableChangesetSpecIDs as _queryPublishableChangesetSpecIDs } from './backend'

const ACTIONS: Action[] = [
    {
        type: 'unpublish',
        buttonLabel: 'Unpublish on apply',
        dropdownTitle: 'Unpublish on apply',
        dropdownDescription:
            'Do not publish selected changesets on the codehost on apply. Note: a changeset that has been published cannot be unpublished.',
        onTrigger: noop,
    },
    {
        type: 'publish',
        buttonLabel: 'Publish on apply',
        dropdownTitle: 'Publish on apply',
        dropdownDescription: 'Publish selected changesets on the codehost on apply.',
        onTrigger: noop,
    },
    {
        type: 'publish-draft',
        buttonLabel: 'Publish draft on apply',
        dropdownTitle: 'Publish draft on apply',
        dropdownDescription: 'Publish selected changesets as drafts on the codehost on apply.',
        onTrigger: noop,
    },
]

// Returns the desired `PublishedValue` for the given action.
const getPublicationStateFromAction = (action: Action): Scalars['PublishedValue'] => {
    switch (action.type) {
        case 'publish':
            return true
        case 'publish-draft':
            return 'draft'
        case 'unpublish':
        default:
            return false
    }
}

export interface PreviewSelectRowProps {
    queryArguments: BatchSpecApplyPreviewVariables
    /** For testing only. */
    dropDownInitiallyOpen?: boolean
    /** For testing only. */
    queryPublishableChangesetSpecIDs?: typeof _queryPublishableChangesetSpecIDs
}

/**
 * Renders the top bar of the PreviewList with the publication state dropdown selector and
 * the X selected label. Provides select ALL functionality.
 */
export const PreviewSelectRow: React.FunctionComponent<PreviewSelectRowProps> = ({
    dropDownInitiallyOpen = false,
    queryPublishableChangesetSpecIDs = _queryPublishableChangesetSpecIDs,
    queryArguments,
}) => {
    const { updatePublicationStates } = useContext(BatchChangePreviewContext)
    const { areAllVisibleSelected, deselectAll, selected, selectAll } = useContext(MultiSelectContext)

    const allChangesetSpecIDs: string[] | undefined = useObservable(
        useMemo(() => queryPublishableChangesetSpecIDs(queryArguments), [
            queryArguments,
            queryPublishableChangesetSpecIDs,
        ])
    )

    const actions = useMemo(
        () =>
            ACTIONS.map(action => {
                const dropdownAction: Action = {
                    ...action,
                    onTrigger: onDone => {
                        updatePublicationStates(
                            [...selected].map(changeSpecID => ({
                                changesetSpec: changeSpecID,
                                publicationState: getPublicationStateFromAction(action),
                            }))
                        )
                        deselectAll()
                        onDone()
                    },
                }

                return dropdownAction
            }),
        [deselectAll, selected, updatePublicationStates]
    )

    return (
        <>
            <div className="row align-items-center no-gutters mb-3">
                <div className="ml-2 col d-flex align-items-center">
                    <InfoCircleOutlineIcon className="icon-inline text-muted mr-2" />
                    {selected === 'all' || allChangesetSpecIDs?.length === selected.size ? (
                        <AllSelectedLabel count={allChangesetSpecIDs?.length} />
                    ) : (
                        `${selected.size} ${pluralize('changeset', selected.size)}`
                    )}
                    {selected !== 'all' &&
                        areAllVisibleSelected() &&
                        allChangesetSpecIDs &&
                        allChangesetSpecIDs.length > selected.size && (
                            <Button className="py-0 px-1" onClick={selectAll} variant="link">
                                (Select all{allChangesetSpecIDs !== undefined && ` ${allChangesetSpecIDs.length}`})
                            </Button>
                        )}
                </div>
                <div className="w-100 d-block d-md-none" />
                <div className="m-0 col col-md-auto">
                    <div className="row no-gutters">
                        <div className="col ml-0 ml-sm-2">
                            <DropdownButton
                                actions={actions}
                                dropdownMenuPosition="right"
                                initiallyOpen={dropDownInitiallyOpen}
                                placeholder="Select action on apply"
                            />
                        </div>
                    </div>
                </div>
            </div>
        </>
    )
}

const AllSelectedLabel: React.FunctionComponent<{ count?: number }> = ({ count }) => {
    if (count === undefined) {
        return <>All changesets selected</>
    }

    return (
        <>
            All {count} {pluralize('changeset', count)} selected
        </>
    )
}
