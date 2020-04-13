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
                <section className="my-3 text-center">
                    {!enableReadAccess && <h1>Tell your admin you are interested in Campaigns</h1>}
                    {enableReadAccess && <h1>How to get started</h1>}

                    <p className="lead">
                        Running Campaigns is currently only supported for site admins of your Sourcegraph instance.{' '}
                        {enableReadAccess && (
                            <>
                                However, your admin can{' '}
                                <a href="https://docs.sourcegraph.com/user/campaigns#configuration" rel="noopener">
                                    enable read-only access to campaigns
                                </a>{' '}
                                for other users.
                            </>
                        )}
                        {!enableReadAccess && (
                            <>
                                Your admin will need to{' '}
                                <a href="https://docs.sourcegraph.com/user/campaigns" rel="noopener">
                                    update the configuration settings
                                </a>{' '}
                                to make large-scale code changes across many repositories and different code hosts.
                            </>
                        )}
                    </p>
                    <div className="row">
                        <ol className="col-6 offset-3 lead text-left">
                            <li>Tell your admin</li>
                            <li>Request read-only access</li>
                        </ol>
                    </div>
                    <div>
                        <button
                            type="button"
                            className="btn btn-primary mr-2"
                            disabled={wasSubmitted}
                            onClick={onUpvote}
                        >
                            {!wasSubmitted && <>Let your admin know you're interested 🖐</>}
                            {wasSubmitted && <>Thanks!</>}
                        </button>
                        <a
                            href="https://docs.sourcegraph.com/user/campaigns"
                            rel="noopener"
                            className="btn btn-primary"
                        >
                            Learn how to get started with campaigns
                        </a>
                    </div>
                </section>
            }
        />
    )
}
