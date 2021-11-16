import classNames from 'classnames'
import ArrowRightIcon from 'mdi-react/ArrowRightIcon'
import React from 'react'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { MarketingBlock } from '@sourcegraph/web/src/components/MarketingBlock'

export interface SelfHostedCtaProps extends TelemetryProps {
    className?: string
    contentClassName?: string
    // the name of the page the CTA will be posted. DO NOT include full URLs
    // here, because this will be logged to our analytics systems. We do not
    // want to expose private repo names or search queries to our analytics.
    page: string
}

export const SelfHostedCta: React.FunctionComponent<SelfHostedCtaProps> = ({
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
                        <a
                            onClick={gettingStartedCTAOnClick}
                            href="https://docs.sourcegraph.com/admin/install"
                            {...linkProps}
                        >
                            Learn how to install
                        </a>
                    </li>
                    <li>
                        <a
                            onClick={selfVsCloudDocumentsLinkOnClick}
                            href="https://docs.sourcegraph.com/code_search/explanations/sourcegraph_cloud#who-is-sourcegraph-cloud-for-why-should-i-use-this-over-sourcegraph-self-hosted"
                            {...linkProps}
                        >
                            Self-hosted vs. cloud features
                        </a>
                    </li>
                </ul>
            </div>

            <MarketingBlock wrapperClassName="flex-md-shrink-0 mt-md-0 mt-sm-2 w-sm-100">
                <h3 className="pr-3">Need help getting started?</h3>

                <div>
                    <a
                        onClick={helpGettingStartedCTAOnClick}
                        href=" https://info.sourcegraph.com/talk-to-a-developer?form_submission_source=inproduct?utm_campaign=inproduct-talktoadev&utm_medium=direct_traffic&utm_source=inproduct-talktoadev&utm_term=null&utm_content=talktoadevform"
                        {...linkProps}
                    >
                        Speak to an engineer
                        <ArrowRightIcon className="icon-inline ml-2" />
                    </a>
                </div>
            </MarketingBlock>
        </div>
    )
}
