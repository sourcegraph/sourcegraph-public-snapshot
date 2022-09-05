import { ButtonLink, LoadingSpinner } from '@sourcegraph/wildcard'
import React from 'react'
import { TrialStartFlowContainer } from '../../TrialStartFlowContainer'
import ClockOutlineIcon from 'mdi-react/ClockOutlineIcon'
import { Link } from 'react-router-dom'

export const WaitForInstancePage: React.FunctionComponent = () => (
    <TrialStartFlowContainer
        afterOutside={
            <p>
                <ButtonLink to="/instances" as={Link} variant="secondary" outline={true} className="text-muted">
                    View more information
                </ButtonLink>
            </p>
        }
    >
        <h3 className="d-flex align-items-center font-weight-normal mb-3">
            <LoadingSpinner className="mr-2" />
            <span>
                Creating <strong>acmelabs.sourcegraph.com</strong>
            </span>
        </h3>
        <p className="text-muted">
            When your new Sourcegraph Cloud instance is ready, you'll get an email at{' '}
            <strong>alice@acme-corp.com</strong>. You can then add repositories and invite other people to start
            searching and navigating code.
        </p>
        <p className="text-muted">
            <ClockOutlineIcon className="icon-inline" /> Less than 1 hour remaining
        </p>
    </TrialStartFlowContainer>
)
