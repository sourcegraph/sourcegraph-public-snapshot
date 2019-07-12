import { NotificationType } from '@sourcegraph/extension-api-classes'
import React, { useCallback, useState } from 'react'
import { ExtensionsControllerProps } from '../../../../../../../shared/src/extensions/controller'
import * as GQL from '../../../../../../../shared/src/graphql/schema'
import { pluralize } from '../../../../../../../shared/src/util/strings'
import { ZapIcon } from '../../../../../util/octicons'
import { ChangesetIcon } from '../../../../changesets/icons'
import { createChangesetFromCodeAction } from '../../../../changesets/preview/backend'
import { ChangesetButtonOrLinkExistingChangeset } from '../../../../tasks/list/item/ChangesetButtonOrLink'
import { ChangesetTargetButtonDropdown } from '../../../../tasks/list/item/ChangesetTargetButtonDropdown'
import { ChangesetPlanProps } from '../useChangesetPlan'

interface Props extends ChangesetPlanProps, ExtensionsControllerProps {
    className?: string
}

const LOADING = 'loading' as const

/**
 * A bar displaying the changesets related to a set of diagnostics, plus a preview of and statistics
 * about a new changeset that is being created.
 */
export const DiagnosticsChangesetsBar: React.FunctionComponent<Props> = ({
    changesetPlan,
    extensionsController,
    className = '',
}) => {
    const [justChanged, setJustChanged] = useState(false)
    const [lastChangesetPlan, setLastChangesetPlan] = useState(changesetPlan)
    if (changesetPlan !== lastChangesetPlan) {
        setJustChanged(true)
        setLastChangesetPlan(changesetPlan)
        setTimeout(() => setJustChanged(false), 500)
    }
    const flashBorderClassName = justChanged ? 'diagnostics-changesets-bar--flash-border' : ''
    const flashBackgroundClassName = justChanged ? 'diagnostics-changesets-bar--flash-bg' : ''

    const [createdThreadOrLoading, setCreatedThreadOrLoading] = useState<ChangesetButtonOrLinkExistingChangeset>(
        LOADING
    )
    const onCreateThreadClick = useCallback(async () => {
        setCreatedThreadOrLoading(PENDING_CREATION)
        try {
            const action = selectedAction
            if (!action) {
                throw new Error('no active code action')
            }
            setCreatedThreadOrLoading(
                await createChangesetFromCodeAction({ extensionsController }, null, action, {
                    status: GQL.ThreadStatus.PREVIEW,
                })
            )
        } catch (err) {
            setCreatedThreadOrLoading(null)
            extensionsController.services.notifications.showMessages.next({
                message: `Error creating changeset: ${err.message}`,
                type: NotificationType.Error,
            })
        }
    }, [extensionsController])

    const isEmpty = changesetPlan.operations.length === 0 || changesetPlan.operations[0].diagnosticActions.length === 0

    return (
        <div className={`diagnostics-changesets-bar ${flashBorderClassName} ${flashBackgroundClassName} ${className}`}>
            <div className="container py-4 d-flex align-items-center">
                <ChangesetTargetButtonDropdown
                    onClick={() => {
                        throw new Error('TODO!(sqs)')
                    }}
                    showAddToExistingChangeset={true}
                    buttonClassName="btn-success"
                    className="mr-3"
                    disabled={isEmpty}
                />

                {!isEmpty ? (
                    <div className={`d-flex align-items-center`}>
                        <span>
                            <ZapIcon className="icon-inline" />{' '}
                            <strong>{changesetPlan.operations[0].diagnosticActions.length}</strong>{' '}
                            <span className="text-muted">
                                {pluralize('action', changesetPlan.operations[0].diagnosticActions.length)}
                            </span>
                        </span>
                    </div>
                ) : (
                    <div className={`text-muted`}>Select actions to include in changeset...</div>
                )}
            </div>
        </div>
    )
}
