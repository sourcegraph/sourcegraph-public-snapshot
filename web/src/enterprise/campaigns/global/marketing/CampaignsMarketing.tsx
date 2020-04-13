import React from 'react'

export interface CampaignsMarketingProps {
    body: JSX.Element
}

export const CampaignsMarketing: React.FunctionComponent<CampaignsMarketingProps> = ({ body }) => (
    <>
        <section className="mt-3 mb-5 text-center">
            <h1 className="font-weight-bold display-4">
                Campaigns <span className="badge badge-info badge-outline">Beta</span>
            </h1>
            <h2 className="mb-6">Make large-scale code changes across all your repositories and code hosts.</h2>
            <div className="position-relative campaign-marketing--video-wrapper">
                <iframe
                    src="https://player.vimeo.com/video/398878670?autoplay=0&title=0&byline=0&portrait=0"
                    className="w-100 h-100 position-absolute campaign-marketing--video-frame"
                    frameBorder="0"
                    allow="autoplay; fullscreen"
                    allowFullScreen={true}
                />
            </div>
            <p className="lead mt-2">
                <a href="https://about.sourcegraph.com/product/code-change-management" rel="noopener">
                    Learn more
                </a>{' '}
                about how campaigns work and what they can do for you.
            </p>
        </section>

        {body}

        <section className="py-5 text-center">
            <h1>Tell us what you think</h1>
            <p className="lead">
                While this feature is still under development and in beta, we appreciate your feedback so we can build
                the best tools for you! Also, <a href="mailto:feedback@sourcegraph.com">reach out to us</a> and let us
                know about more feedback.
            </p>
        </section>
    </>
)
