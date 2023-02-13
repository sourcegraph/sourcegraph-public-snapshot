import React from 'react'

import { mdiOpenInNew } from '@mdi/js'

import { AuthenticatedUser } from '@sourcegraph/shared/src/auth'
import { buildCloudTrialURL } from '@sourcegraph/shared/src/util/url'
import { Container, H2, H3, Link, Text, Icon } from '@sourcegraph/wildcard'

import { BatchChangesIcon } from '../../../batches/icons'
import { CloudCtaBanner } from '../../../components/CloudCtaBanner'
import { CtaBanner } from '../../../components/CtaBanner'
import { eventLogger } from '../../../tracking/eventLogger'

export interface GettingStartedProps {
    isSourcegraphDotCom: boolean
    authenticatedUser?: Pick<AuthenticatedUser, 'displayName' | 'emails'> | null
    className?: string
}

export const GettingStarted: React.FunctionComponent<React.PropsWithChildren<GettingStartedProps>> = ({
    isSourcegraphDotCom,
    authenticatedUser,
    className,
}) => (
    <div className={className} data-testid="test-getting-started">
        <Container className="mb-3">
            <div className="row align-items-center">
                <div className="col-12 col-md-7">
                    <video
                        className="w-100 h-auto shadow percy-hide"
                        width={1280}
                        height={720}
                        autoPlay={true}
                        muted={true}
                        loop={true}
                        playsInline={true}
                        controls={false}
                    >
                        <source
                            type="video/webm"
                            src="https://storage.googleapis.com/sourcegraph-assets/batch-changes/how-it-works.webm"
                        />
                        <source
                            type="video/mp4"
                            src="https://storage.googleapis.com/sourcegraph-assets/batch-changes/how-it-works.mp4"
                        />
                    </video>
                </div>
                <div className="col-12 col-md-5">
                    <H2>Automate large-scale code changes</H2>
                    <Text>
                        Batch Changes makes it easy to find and change code across many repositories (or many subtrees
                        in a big monorepo). It lets you create, update, and track pull requests to ensure the change is
                        reviewed, tested, and safely merged everywhere.
                    </Text>
                    <H3>Use Batch Changes to...</H3>
                    <ul>
                        <li>Update configuration files across many repositories</li>
                        <li>Update libraries consuming your APIs</li>
                        <li>Rapidly fix critical security issues</li>
                        <li>Update boilerplate code</li>
                        <li>Pay down tech debt</li>
                    </ul>
                    <H3>Resources</H3>
                    <ul>
                        <li>
                            <Link to="/help/batch_changes" target="_blank" rel="noopener">
                                Documentation <Icon role="img" aria-label="Open in a new tab" svgPath={mdiOpenInNew} />
                            </Link>
                        </li>
                        <li>
                            <Link to="https://about.sourcegraph.com/batch-changes" target="_blank" rel="noopener">
                                Product page <Icon role="img" aria-label="Open in a new tab" svgPath={mdiOpenInNew} />
                            </Link>
                        </li>
                    </ul>
                </div>
            </div>
        </Container>
        {isSourcegraphDotCom ? (
            <CloudCtaBanner variant="filled">
                To automate changes across your team's private repos,{' '}
                <Link
                    to={buildCloudTrialURL(authenticatedUser, 'batch')}
                    target="_blank"
                    rel="noopener noreferrer"
                    onClick={() => eventLogger.log('ClickedOnCloudCTA', { cloudCtaType: 'BatchChangesGettingStarted' })}
                >
                    try Sourcegraph Cloud
                </Link>
                .
            </CloudCtaBanner>
        ) : (
            <div className="d-flex justify-content-start">
                <CtaBanner
                    bodyText="Try it yourself in less than 10 minutes (without actually pushing changes)."
                    title={<H3>Start using Batch Changes</H3>}
                    linkText="Read quickstart docs"
                    href="/help/batch_changes/quickstart"
                    icon={<BatchChangesIcon />}
                />
            </div>
        )}
    </div>
)
