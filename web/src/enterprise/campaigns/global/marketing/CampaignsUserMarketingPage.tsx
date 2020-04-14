import React, { useState } from 'react'
import { CampaignsMarketing } from './CampaignsMarketing'
import { TelemetryProps } from '../../../../../../shared/src/telemetry/telemetryService'

export interface CampaignsUserMarketingPageProps extends TelemetryProps {
    /** When true, the message shown will indicate that the user can just not use it because read-access has not been provided to non site-admins. */
    enableReadAccess: boolean
}

export const CampaignsUserMarketingPage: React.FunctionComponent<CampaignsUserMarketingPageProps> = ({
    enableReadAccess,
    telemetryService,
}) => {
    const [wasSubmitted, setWasSubmitted] = useState<boolean>(false)
    const onUpvote: React.MouseEventHandler = () => {
        telemetryService.log('ExpressCampaignsInterest')
        setWasSubmitted(true)
    }
    return (
        <CampaignsMarketing
            body={
                <section className="my-3">
                    <h2>Interested in Campaigns?</h2>
                    <p>
                        At this time, creating and managing campaigns is only available to Sourcegraph admins. What can
                        you do?
                    </p>
                    <div className="row">
                        <ol>
                            <li>Let your Sourcegraph admin know you're interested in using Campaigns for your team.</li>
                            {enableReadAccess && (
                                <li>
                                    Ask your Sourcegraph admin for{' '}
                                    <a href="https://docs.sourcegraph.com/user/campaigns#configuration" rel="noopener">
                                        read-only access
                                    </a>{' '}
                                    to Campaigns.
                                    <div className="alert alert-info mt-3">
                                        <b>NOTE:</b> Repository permissions are NOT enforced by campaigns, so your admin
                                        may not grant read-only access if your Sourcegraph instance has repository
                                        permissions configured.
                                    </div>
                                </li>
                            )}
                            <li>
                                Learn how to{' '}
                                <a href="https://docs.sourcegraph.com/user/campaigns#creating-campaigns">
                                    get started creating campaigns
                                </a>
                                .
                            </li>
                        </ol>
                    </div>

                    <div>
                        <button
                            type="button"
                            className="btn btn-primary mr-2"
                            disabled={wasSubmitted}
                            onClick={onUpvote}
                        >
                            {!wasSubmitted && <>Notify I'm interested 🖐</>}
                            {wasSubmitted && <>Thanks!</>}
                        </button>
                        <a
                            href="https://docs.sourcegraph.com/user/campaigns"
                            rel="noopener"
                            className="btn btn-primary"
                        >
                            Learn how to get started
                        </a>
                    </div>
                </section>
            }
        />
    )
}
