import React from 'react'
import { CampaignsMarketing } from './CampaignsMarketing'
import { TelemetryProps } from '../../../../../../shared/src/telemetry/telemetryService'

export interface CampaignsDotComPageProps extends TelemetryProps {}

export const CampaignsDotComPage: React.FunctionComponent<CampaignsDotComPageProps> = ({ telemetryService }) => (
    <>
        <CampaignsMarketing
            telemetryService={telemetryService}
            subTitle={
                <i>
                    Campaigns are not available on Sourcegraph.com
                    <br />
                    <a href="https://docs.sourcegraph.com/admin/install" rel="noopener">
                        Create a free private instance
                    </a>{' '}
                    to get started.
                </i>
            }
            showFeedbackUI={false}
        />

        <section className="my-3 text-center">
            <h1>Start using campaigns</h1>

            <p className="lead">
                <a href="" rel="noopener">
                    Install Sourcegraph locally
                </a>{' '}
                with a free private instance and{' '}
                <a href="https://docs.sourcegraph.com/user/campaigns" rel="noopener">
                    update the configuration settings
                </a>{' '}
                to use campaigns.
            </p>

            <a href="https://docs.sourcegraph.com/admin/install" rel="noopener" className="btn btn-primary">
                Get started now
            </a>
        </section>
    </>
)
