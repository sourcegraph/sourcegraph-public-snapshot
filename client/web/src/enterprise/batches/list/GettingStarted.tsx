import React from 'react'

import { mdiOpenInNew } from '@mdi/js'

import { DownloadSourcegraphIcon } from '@sourcegraph/branded/src/components/DownloadSourcegraphIcon'
import { Container, Card, Link, H2, H3, H4, Text, Icon } from '@sourcegraph/wildcard'

import { BatchChangesIcon } from '../../../batches/icons'
import { CtaBanner } from '../../../components/CtaBanner'
import { eventLogger } from '../../../tracking/eventLogger'

export interface GettingStartedProps {
    isSourcegraphDotCom: boolean
    className?: string
}

export const GettingStarted: React.FunctionComponent<React.PropsWithChildren<GettingStartedProps>> = ({
    isSourcegraphDotCom,
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
                                Documentation{' '}
                                <Icon role="img" aria-label="Open in a new tab" svgPath={mdiOpenInNew} />
                            </Link>
                        </li>
                        <li>
                            <Link to="https://about.sourcegraph.com/batch-changes" target="_blank" rel="noopener">
                                Product page{' '}
                                <Icon role="img" aria-label="Open in a new tab" svgPath={mdiOpenInNew} />
                            </Link>
                        </li>
                    </ul>
                    
                    {isSourcegraphDotCom && (
                        <Card className="shadow d-flex flex-row align-items-center p-3 mt-2">
                            <Icon as={BatchChangesIcon} violetBg={true} role="img" size="md" aria-hidden={true} />
                            <div className="pl-3">
                                <H4 className="mb-1">Get Batch Changes on your code</H4>
                                <Link to="https://signup.sourcegraph.com/" onClick={() => eventLogger.log('ClickedOnCloudCTA', { url: window.location.href })}>
                                    Sign up for a 30-day trial on Sourcegraph Cloud.
                                </Link>
                            </div>
                        </Card>
                    )}
                </div>
            </div>
        </Container>
        <div className="d-flex justify-content-start">
            {isSourcegraphDotCom ? (
                <CtaBanner
                    className="col-lg-7"
                    bodyText="Batch Changes requires a Sourcegraph Cloud or self-hosted instance."
                    title={<H3>Install locally</H3>}
                    linkText="Install local instance"
                    href="/help/admin/install?utm_medium=inproduct&utm_source=inproduct-batch-changes&utm_campaign=inproduct-batch-changes&term="
                    icon={<DownloadSourcegraphIcon />}
                    onClick={() => eventLogger.log('BatchChangesInstallFromCloudClicked')}
                />
            ) : (
                <CtaBanner
                    bodyText="Try it yourself in less than 10 minutes (without actually pushing changes)."
                    title={<H3>Start using Batch Changes</H3>}
                    linkText="Read quickstart docs"
                    href="/help/batch_changes/quickstart"
                    icon={<BatchChangesIcon />}
                />
            )}
        </div>
    </div>
)
