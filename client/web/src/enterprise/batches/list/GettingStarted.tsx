import React from 'react'

import classNames from 'classnames'

import { Container, CardBody, Card, Link, Typography } from '@sourcegraph/wildcard'

import styles from './GettingStarted.module.scss'

export interface GettingStartedProps {
    footer?: React.ReactNode
    className?: string
}

export const GettingStarted: React.FunctionComponent<React.PropsWithChildren<GettingStartedProps>> = ({
    footer,
    className,
}) => (
    <div className={className} data-testid="test-getting-started">
        <Container className="mb-4">
            <div className={classNames(styles.videoRow, 'row')}>
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
                    <Typography.H2>Automate large-scale code changes</Typography.H2>
                    <p>
                        Batch Changes gives you a declarative structure for finding and modifying code across all of
                        your repositories. Its simple UI makes it easy to track and manage all of your changesets
                        through checks and code reviews until each change is merged.
                    </p>
                    <Typography.H3>Common use cases</Typography.H3>
                    <ul className={classNames(styles.narrowList, 'mb-0')}>
                        <li>Update configuration files across many repositories</li>
                        <li>Update libraries which call your APIs</li>
                        <li>Rapidly fix critical security issues</li>
                        <li>Update boilerplate code</li>
                        <li>Pay down tech debt</li>
                    </ul>
                </div>
            </div>
        </Container>
        <Typography.H3 className="mb-3">Tutorials to help with your first batch change</Typography.H3>
        <div className="row">
            <div className="col-12 col-md-6 mb-2">
                <Card className="h-100">
                    <CardBody className="d-flex">
                        <FindReplaceIcon className="mr-3" />
                        <div>
                            <Typography.H4>
                                <Link
                                    to="/help/batch_changes/tutorials/search_and_replace_specific_terms"
                                    rel="noopener"
                                >
                                    Finding and replacing exclusionary terms
                                </Link>
                            </Typography.H4>
                            <p className="text-muted mb-0">
                                A Sourcegraph query plus a simple <code>sed</code> command creates changesets required
                                to manage a large scale change.
                            </p>
                        </div>
                    </CardBody>
                </Card>
            </div>
            <div className="col-12 col-md-6 mb-3">
                <Card className="h-100">
                    <CardBody className="d-flex">
                        <RefactorCombyIcon className="mr-3" />
                        <div>
                            <Typography.H4>
                                <Link to="/help/batch_changes/tutorials/updating_go_import_statements" rel="noopener">
                                    Refactoring with language aware search
                                </Link>
                            </Typography.H4>
                            <p className="text-muted mb-0">
                                Using{' '}
                                <Link to="https://comby.dev/" rel="noopener">
                                    Comby's
                                </Link>{' '}
                                language-aware structural search to refactor Go statements to a semantically equivalent,
                                but clearer execution.
                            </p>
                        </div>
                    </CardBody>
                </Card>
            </div>
            <div className="col-12 mb-4 text-right">
                <p>
                    <Link to="/help/batch_changes/tutorials" rel="noopener">
                        More tutorials
                    </Link>
                </p>
            </div>
        </div>

        <div className="row mb-5">
            <div className="col-12 col-md-4 mt-3">
                <p>
                    <strong>Quickstart</strong>
                </p>
                <p>Create your first Sourcegraph batch change in 10 minutes or less.</p>
                <Link to="/help/batch_changes/quickstart" rel="noopener">
                    Batch Changes quickstart
                </Link>
            </div>
            <div className="col-12 col-md-4 mt-3">
                <p>
                    <strong>Documentation</strong>
                </p>
                <p>
                    Learn about the batch spec{' '}
                    <Link to="/help/batch_changes/references/batch_spec_yaml_reference" rel="noopener">
                        YAML reference
                    </Link>
                    , its powerful{' '}
                    <Link to="/help/batch_changes/references/batch_spec_templating">templating language</Link>,{' '}
                    <Link to="/help/batch_changes/explanations/permissions_in_batch_changes" rel="noopener">
                        permissions
                    </Link>{' '}
                    and more in the{' '}
                    <Link to="/help/batch_changes" rel="noopener">
                        Batch Changes documentation
                    </Link>
                    .
                </p>
            </div>
            <div className="col-12 col-md-4">
                <Card className={styles.overviewCard}>
                    <CardBody>
                        <p>
                            <strong>Overview</strong>
                        </p>
                        <p>
                            View the product marketing page for a high-level overview of benefits and customer use
                            cases.
                        </p>
                        {/*
                            a11y-ignore
                            Rule: "color-contrast" (Elements must have sufficient color contrast)
                            GitHub issue: https://github.com/sourcegraph/sourcegraph/issues/33343
                        */}
                        <Link to="https://about.sourcegraph.com/batch-changes" rel="noopener" className="a11y-ignore">
                            Batch Changes marketing page
                        </Link>
                    </CardBody>
                </Card>
            </div>
        </div>
        <Typography.H2>Batch changes demo</Typography.H2>
        <p>
            This 2 minute demo provides an overview of batch changes from editing a specification to managing
            changesets.
        </p>
        <Container className="mb-3">
            <iframe
                title="Batch Changes demo"
                className="percy-hide chromatic-ignore"
                width="100%"
                height="600"
                src="https://www.youtube-nocookie.com/embed/eOmiyXIWTCw"
                frameBorder="0"
                allow="accelerometer; autoplay; encrypted-media; gyroscope; picture-in-picture"
                allowFullScreen={true}
            />
        </Container>
        {footer}
    </div>
)

const FindReplaceIcon: React.FunctionComponent<React.PropsWithChildren<{ className?: string }>> = ({ className }) => (
    <svg
        width="103"
        height="60"
        viewBox="0 0 103 60"
        fill="none"
        className={className}
        xmlns="http://www.w3.org/2000/svg"
        role="presentation"
    >
        <rect width="54" height="8" rx="4" fill="#CAD2E2" />
        <rect x="57" width="32" height="8" rx="4" fill="#CAD2E2" />
        <rect x="92" width="11" height="8" rx="4" fill="#CAD2E2" />
        <rect x="10" y="13" width="88" height="8" rx="4" fill="#CAD2E2" />
        <rect x="10" y="26" width="38" height="8" rx="4" fill="#CAD2E2" />
        <rect x="51" y="26" width="38" height="8" rx="4" fill="#F03E3E" />
        <g filter="url(#filter0_d)">
            <rect x="57" y="18" width="40" height="9" rx="4.5" fill="#94D82D" />
        </g>
        <rect x="9.99792" y="39" width="25" height="8" rx="4" fill="#CAD2E2" />
        <rect y="52" width="11" height="8" rx="4" fill="#CAD2E2" />
        <defs>
            <filter
                id="filter0_d"
                x="54.9282"
                y="18"
                width="44.1436"
                height="13.1436"
                filterUnits="userSpaceOnUse"
                colorInterpolationFilters="sRGB"
            >
                <feFlood floodOpacity="0" result="BackgroundImageFix" />
                <feColorMatrix
                    in="SourceAlpha"
                    type="matrix"
                    values="0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 127 0"
                    result="hardAlpha"
                />
                <feOffset dy="2.07182" />
                <feGaussianBlur stdDeviation="1.03591" />
                <feColorMatrix type="matrix" values="0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0.25 0" />
                <feBlend mode="normal" in2="BackgroundImageFix" result="effect1_dropShadow" />
                <feBlend mode="normal" in="SourceGraphic" in2="effect1_dropShadow" result="shape" />
            </filter>
        </defs>
    </svg>
)

const RefactorCombyIcon: React.FunctionComponent<React.PropsWithChildren<{ className?: string }>> = ({ className }) => (
    <svg
        width="111"
        height="60"
        viewBox="0 0 111 60"
        fill="none"
        className={className}
        xmlns="http://www.w3.org/2000/svg"
        role="presentation"
    >
        <rect width="54" height="8" rx="4" fill="#CAD2E2" />
        <rect x="57" width="32" height="8" rx="4" fill="#CAD2E2" />
        <rect x="92" width="11" height="8" rx="4" fill="#CAD2E2" />
        <rect x="9.99792" y="13" width="14" height="8" rx="4" fill="#F91624" />
        <rect x="29.4938" y="13" width="38" height="8" rx="4" fill="#F91624" />
        <rect x="72.6543" y="13" width="21" height="8" rx="4" fill="#CAD2E2" />
        <rect x="99.1487" y="13" width="11" height="8" rx="4" fill="#F91624" />
        <rect x="9.99792" y="26" width="38" height="8" rx="4" fill="#94D82D" />
        <rect x="51" y="26" width="21" height="8" rx="4" fill="#CAD2E2" />
        <rect x="75" y="26" width="11" height="8" rx="4" fill="#94D82D" />
        <rect x="9.99792" y="39" width="25" height="8" rx="4" fill="#CAD2E2" />
        <rect y="52" width="11" height="8" rx="4" fill="#CAD2E2" />
    </svg>
)
