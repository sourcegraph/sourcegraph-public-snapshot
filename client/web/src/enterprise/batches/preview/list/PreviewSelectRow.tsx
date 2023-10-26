import React, { useMemo, useContext } from 'react'

import { mdiInformationOutline } from '@mdi/js'
import { noop } from 'lodash'

import { pluralize } from '@sourcegraph/common'
import { Button, useObservable, Icon } from '@sourcegraph/wildcard'

import type { BatchSpecApplyPreviewVariables, Scalars } from '../../../../graphql-operations'
import { type Action, DropdownButton } from '../../DropdownButton'
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
        case 'publish': {
            return true
        }
        case 'publish-draft': {
            return 'draft'
        }
        case 'unpublish':
        default: {
            return false
        }
    }
}

export interface PreviewSelectRowProps {
    queryArguments: BatchSpecApplyPreviewVariables
    /** For testing only. */
    queryPublishableChangesetSpecIDs?: typeof _queryPublishableChangesetSpecIDs
}

/**
 * Renders the top bar of the PreviewList with the publication state dropdown selector and
 * the X selected label. Provides select ALL functionality.
 */
export const PreviewSelectRow: React.FunctionComponent<React.PropsWithChildren<PreviewSelectRowProps>> = ({
    queryPublishableChangesetSpecIDs = _queryPublishableChangesetSpecIDs,
    queryArguments,
}) => {
    // The user can modify the desired publication states for changesets in the preview
    // list from this dropdown selector. However, these modifications are transient and
    // are not persisted to the backend (until the user applies the batch change and the
    // publication states are realized, of course). Rather, they are provided as arguments
    // to the `applyPreview` connection, and later the `applyBatchChange` mutation, in
    // order to override the original publication states computed by the reconciler on the
    // backend. `BatchChangePreviewContext` is responsible for managing these publication
    // states clientside.
    const { updatePublicationStates } = useContext(BatchChangePreviewContext)
    const { areAllVisibleSelected, deselectAll, selected, selectAll } = useContext(MultiSelectContext)

    const allChangesetSpecIDs: string[] | undefined = useObservable(
        useMemo(
            () => queryPublishableChangesetSpecIDs(queryArguments),
            [queryArguments, queryPublishableChangesetSpecIDs]
        )
    )

    const actions = useMemo(
        () =>
            ACTIONS.map(action => {
                const dropdownAction: Action = {
                    ...action,
                    onTrigger: onDone => {
                        const specIDs = selected === 'all' ? allChangesetSpecIDs : [...selected]
                        if (!specIDs) {
                            // allChangesetSpecIDs hasn't populated yet: it
                            // shouldn't be possible to set selected to 'all' if
                            // that's the case, but to be safe, we'll just bail
                            // early if that somehow happens.
                            return
                        }

                        // Record the new desired publication state for each selected changeset.
                        updatePublicationStates(
                            specIDs.map(changeSpecID => ({
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
        [allChangesetSpecIDs, deselectAll, selected, updatePublicationStates]
    )

    return (
        <>
            <div className="row align-items-center no-gutters mb-3">
                <div className="ml-2 col d-flex align-items-center">
                    <Icon aria-hidden={true} className="text-muted mr-2" svgPath={mdiInformationOutline} />
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
                            <DropdownButton actions={actions} placeholder="Select action on apply" />
                        </div>
                    </div>
                </div>
            </div>
        </>
    )
}

const AllSelectedLabel: React.FunctionComponent<React.PropsWithChildren<{ count?: number }>> = ({ count }) => {
    if (count === undefined) {
        return <>All changesets selected</>
    }

    return (
        <>
            All {count} {pluralize('changeset', count)} selected
        </>
    )
}
