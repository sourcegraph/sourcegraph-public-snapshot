import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import * as GQL from '../../../../../shared/src/graphql/schema'
import H from 'history'
import React, { useMemo } from 'react'
import { ExtensionsControllerProps } from '../../../../../shared/src/extensions/controller'
import { PlatformContextProps } from '../../../../../shared/src/platform/context'
import { isErrorLike } from '../../../../../shared/src/util/errors'
import { ThemeProps } from '../../../theme'
import { CampaignFormData } from '../form/CampaignForm'
import { useCampaignUpdatePreview } from './useCampaignUpdatePreview'
import { isDefined } from '../../../../../shared/src/util/types'
import { Timestamp } from '../../../components/time/Timestamp'
import { ThreadUpdatePreviewList } from '../../threads/updatePreview/ThreadUpdatePreviewList'

interface Props extends ExtensionsControllerProps, PlatformContextProps, ThemeProps {
    campaign: Pick<GQL.ICampaign, 'id'>
    data: CampaignFormData

    className?: string
    location: H.Location
    history: H.History
}

const LOADING = 'loading' as const

interface PropertyUpdate {
    name: keyof GQL.ICampaign
    displayName: string
    oldValue: React.ReactFragment
    newValue: React.ReactFragment
}

/**
 * A campaign update preview.
 */
export const CampaignUpdatePreview: React.FunctionComponent<Props> = ({ campaign, data, className = '', ...props }) => {
    const [preview, isLoading] = useCampaignUpdatePreview(props, { id: campaign.id, ...data })

    const propertyUpdates = useMemo<PropertyUpdate[]>(
        () =>
            preview !== LOADING && !isErrorLike(preview)
                ? [
                      preview.oldName !== null && preview.newName !== null && preview.newName !== preview.oldName
                          ? {
                                name: 'name' as const,
                                displayName: 'Name',
                                oldValue: preview.oldName,
                                newValue: preview.newName,
                            }
                          : undefined,
                      preview.oldStartDate !== preview.newStartDate
                          ? {
                                name: 'startDate' as const,
                                displayName: 'Start date',
                                oldValue: preview.oldStartDate ? <Timestamp date={preview.oldStartDate} /> : 'none',
                                newValue: preview.newStartDate ? <Timestamp date={preview.newStartDate} /> : 'none',
                            }
                          : undefined,
                      preview.oldDueDate !== preview.newDueDate
                          ? {
                                name: 'dueDate' as const,
                                displayName: 'Due date',
                                oldValue: preview.oldDueDate ? <Timestamp date={preview.oldDueDate} /> : 'none',
                                newValue: preview.newDueDate ? <Timestamp date={preview.newDueDate} /> : 'none',
                            }
                          : undefined,
                  ].filter(isDefined)
                : [],
        [preview]
    )

    return (
        <div className={`campaign-preview ${className}`}>
            <h2 className="d-flex align-items-center">
                Preview
                {isLoading && <LoadingSpinner className="icon-inline ml-3" />}
            </h2>
            {preview !== LOADING &&
                (isErrorLike(preview) ? (
                    <div className="alert alert-danger">Error: {preview.message}</div>
                ) : (
                    // eslint-disable-next-line react/forbid-dom-props
                    <div style={isLoading ? { opacity: 0.5, cursor: 'wait' } : undefined}>
                        {propertyUpdates.length === 0 && (!preview.threads || preview.threads.length === 0) && (
                            <p className="text-muted">No changes</p>
                        )}
                        {propertyUpdates.length > 0 && (
                            <ul>
                                {propertyUpdates.map(({ name, displayName, oldValue, newValue }) => (
                                    <li key={name}>
                                        <strong>{displayName}:</strong> changed to <strong>{newValue}</strong> from{' '}
                                        <s>{oldValue}</s>
                                    </li>
                                ))}
                            </ul>
                        )}
                        {preview.threads && preview.threads.length > 0 && (
                            <>
                                <a id="threads" />
                                <ThreadUpdatePreviewList
                                    {...props}
                                    threadUpdatePreviews={preview.threads}
                                    showRepository={true}
                                    headerItems={{
                                        left: <h4 className="mb-0">Changesets &amp; issues</h4>,
                                    }}
                                    className="mb-4"
                                />
                            </>
                        )}
                    </div>
                ))}
        </div>
    )
}
