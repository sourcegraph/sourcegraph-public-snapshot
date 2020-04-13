import React, { useState } from 'react'
import GithubCircleIcon from 'mdi-react/GithubCircleIcon'
import GitlabIcon from 'mdi-react/GitlabIcon'
import BitbucketIcon from 'mdi-react/BitbucketIcon'
import CheckCircleIcon from 'mdi-react/CheckCircleIcon'
import { TelemetryProps } from '../../../../../../shared/src/telemetry/telemetryService'

export interface CampaignsMarketingProps extends TelemetryProps {
    showFeedbackUI?: boolean
    subTitle: JSX.Element
}

export const CampaignsMarketing: React.FunctionComponent<CampaignsMarketingProps> = ({
    subTitle,
    showFeedbackUI = true,
    telemetryService,
}) => {
    const [wasSubmitted, setWasSubmitted] = useState<boolean>(false)
    const onRate = (rating: string): React.MouseEventHandler => () => {
        telemetryService.log('rate_campaigns', { rating })
        setWasSubmitted(true)
    }
    return (
        <>
            <section className="mt-3 text-center">
                <h1 className="display-1 font-weight-bold">
                    Large scale
                    <br />
                    code changes
                </h1>
                <h1>
                    <span className="badge badge-info badge-outline">Public beta</span>
                </h1>
                <h1 className="mb-6">{subTitle}</h1>
                <div className="position-relative mb-5 campaign-marketing--video-wrapper">
                    <iframe
                        src="https://player.vimeo.com/video/398878670?autoplay=0&title=0&byline=0&portrait=0"
                        className="w-100 h-100 position-absolute campaign-marketing--video-frame"
                        frameBorder="0"
                        allow="autoplay; fullscreen"
                        allowFullScreen={true}
                    />
                </div>
            </section>

            <section className="mb-3">
                <h1 className="text-center mb-5">Code change campaigns help teams move quickly and safely</h1>
                <div className="row">
                    <div className="col-12 col-md-6">
                        <h2>Remove deprecated code and legacy systems</h2>
                        <p>
                            You need a way to improve and change APIs used across all of your organization's code, to
                            spend less time and effort in the migration period between the old and new API or service.
                        </p>
                    </div>
                    <div className="col-12 col-md-6">
                        <h2>Keep dependencies up-to-date</h2>
                        <p>
                            Keep your library dependencies and how you use those libraries up-to-date and consistent
                            across all of your organization's code, to avoid old bugs or security problems in old
                            dependencies, and problems arising from inconsistent dependency version use across your
                            codebase.
                        </p>
                    </div>
                </div>
                <div className="row">
                    <div className="col-12 col-md-6">
                        <h2>Deploy new static analysis gradually in the developer workflow</h2>
                        <p>
                            Increase adoption of linters and enable progressively stricter rules across all of your
                            organization's code, so you can continuously improve the quality of all of your code.
                            Developers will see diagnostics and fixes in their editor and on their code host to gently
                            nudge them toward adherence, and you can enforce rules after a certain time period.
                        </p>
                    </div>
                    <div className="col-12 col-md-6">
                        <h2>Triage and follow-through on critical security issues</h2>
                        <p>
                            You need to be able to identify everywhere that a vulnerable package or API is used, and
                            open issues or pull requests on all affected projects. Then you can monitor the progress of
                            fixing, merging, and deploying.
                        </p>
                    </div>
                </div>
                <div className="row">
                    <div className="col-12 col-md-6">
                        <h2>Standardize build and deploy configuration</h2>
                        <p>
                            Keep the build and deployment configurations up-to-date and consistent across all of your
                            organization's code, so that you can iterate and deploy continuously and reliably with
                            DevOps self-sufficiency.
                        </p>
                    </div>
                </div>
            </section>

            <section className="py-3 text-center">
                <h1>Integrates with your favorite code host</h1>
                <p className="lead">Campaigns are currently supporting the following code hosts:</p>

                <div className="d-flex justify-content-around mt-4">
                    <div className="flex-grow-0">
                        <div className="text-center">
                            <GithubCircleIcon size="4rem" />
                            <h3>GitHub.com / Enterprise</h3>
                            <p className="mb-1 text-left">
                                <CheckCircleIcon className="icon-inline text-success" /> Create and track pull requests
                            </p>
                            <p className="mb-1 text-left">
                                <CheckCircleIcon className="icon-inline text-success" /> Review status on pull requests
                            </p>
                            <p className="mb-1 text-left">
                                <CheckCircleIcon className="icon-inline text-success" /> CI status on pull requests
                            </p>
                            <p className="mb-1 text-left">
                                <CheckCircleIcon className="icon-inline text-success" /> Webhook support
                            </p>
                        </div>
                    </div>
                    <div className="flex-grow-0">
                        <div className="text-center">
                            <BitbucketIcon size="4rem" />
                            <h3>Bitbucket Server</h3>
                            <p className="mb-1 text-left">
                                <CheckCircleIcon className="icon-inline text-success" /> Create and track pull requests
                            </p>
                            <p className="mb-1 text-left">
                                <CheckCircleIcon className="icon-inline text-success" /> Review status on pull requests
                            </p>
                            <p className="mb-1 text-left">
                                <CheckCircleIcon className="icon-inline text-success" /> CI status on pull requests
                            </p>
                            <p className="mb-1 text-left">
                                <CheckCircleIcon className="icon-inline text-success" /> Webhook support
                            </p>
                        </div>
                    </div>
                    <div className="flex-grow-0">
                        <div className="text-center">
                            <GitlabIcon size="4rem" />
                            <h3>Gitlab.com / Enterprise</h3>
                            <p className="mb-0 text-muted">Coming soon!</p>
                        </div>
                    </div>
                </div>
            </section>

            {showFeedbackUI && (
                <section className="py-5 text-center">
                    <h1>Tell us what you think</h1>
                    <p className="lead">
                        While this feature is still under development and in public beta, we appreciate your feedback so
                        we can build the best tools for you!
                    </p>
                    <div className="d-flex mt-5 align-items-stretch justify-content-around">
                        <div>
                            <button
                                type="button"
                                className="btn campaign-marketing--rate-btn"
                                disabled={wasSubmitted}
                                onClick={onRate('+1')}
                            >
                                👍
                            </button>
                        </div>
                        <div>
                            <button
                                type="button"
                                className="btn campaign-marketing--rate-btn"
                                disabled={wasSubmitted}
                                onClick={onRate('rocket')}
                            >
                                🚀
                            </button>
                        </div>
                        <div>
                            <button
                                type="button"
                                className="btn campaign-marketing--rate-btn"
                                disabled={wasSubmitted}
                                onClick={onRate('thinking')}
                            >
                                🤔
                            </button>
                        </div>
                        <div>
                            <button
                                type="button"
                                className="btn campaign-marketing--rate-btn"
                                disabled={wasSubmitted}
                                onClick={onRate('-1')}
                            >
                                👎
                            </button>
                        </div>
                    </div>
                </section>
            )}
        </>
    )
}
