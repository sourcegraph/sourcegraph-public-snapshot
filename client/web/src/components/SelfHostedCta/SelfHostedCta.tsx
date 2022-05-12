import React from 'react'

import classNames from 'classnames'
import ArrowRightIcon from 'mdi-react/ArrowRightIcon'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Link, Icon, Typography } from '@sourcegraph/wildcard'

import { MarketingBlock } from '../MarketingBlock'

export interface SelfHostedCtaProps extends TelemetryProps {
    className?: string
    contentClassName?: string
    // the name of the page the CTA will be posted. DO NOT include full URLs
    // here, because this will be logged to our analytics systems. We do not
    // want to expose private repo names or search queries to our analytics.
    page: string
}

export const SelfHostedCta: React.FunctionComponent<React.PropsWithChildren<SelfHostedCtaProps>> = ({
    className,
    contentClassName,
    telemetryService,
    page,
    children,
}) => {
    const linkProps = { rel: 'noopener noreferrer' }

    const gettingStartedCTAOnClick = (): void => {
        telemetryService.log('InstallSourcegraphCTAClicked', { page }, { page })
    }

    const selfVsCloudDocumentsLinkOnClick = (): void => {
        telemetryService.log('SelfVsCloudDocsLink', { page }, { page })
    }

    const helpGettingStartedCTAOnClick = (): void => {
        telemetryService.log('HelpGettingStartedCTA', { page }, { page })
    }

    return (
        <div
            className={classNames(
                'd-flex flex-md-row align-items-md-start justify-content-md-between flex-column',
                className
            )}
        >
            <div className={classNames('mr-md-4 mr-0', contentClassName)}>
                {children}

                <ul>
                    <li>
                        <Link
                            onClick={gettingStartedCTAOnClick}
                            to="https://docs.sourcegraph.com/admin/install"
                            {...linkProps}
                        >
                            Learn how to install
                        </Link>
                    </li>
                    <li>
                        <Link
                            onClick={selfVsCloudDocumentsLinkOnClick}
                            to="https://docs.sourcegraph.com/code_search/explanations/sourcegraph_cloud#who-is-sourcegraph-cloud-for-why-should-i-use-this-over-sourcegraph-self-hosted"
                            {...linkProps}
                        >
                            Self-hosted vs. cloud features
                        </Link>
                    </li>
                </ul>
            </div>

            <MarketingBlock wrapperClassName="flex-md-shrink-0 mt-md-0 mt-sm-2 w-sm-100">
                <Typography.H3 className="pr-3">Need help getting started?</Typography.H3>

                <div>
                    <Link
                        onClick={helpGettingStartedCTAOnClick}
                        to="https://info.sourcegraph.com/talk-to-a-developer?form_submission_source=inproduct&utm_campaign=inproduct-talktoadev&utm_medium=direct_traffic&utm_source=in-product&utm_term=null&utm_content=talktoadevform"
                        {...linkProps}
                    >
                        Speak to an engineer
                        <Icon className="ml-2" as={ArrowRightIcon} />
                    </Link>
                </div>
            </MarketingBlock>
        </div>
    )
}
