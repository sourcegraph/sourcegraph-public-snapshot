import React, { useEffect, useState } from 'react'

import classNames from 'classnames'
import { noop } from 'lodash'
import { Navigate, useLocation } from 'react-router-dom-v5-compat'

import { Text, Link, ErrorAlert, Form, Input, Button, LoadingSpinner, TextArea } from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../auth'
import { HeroPage } from '../components/HeroPage'
import { PageTitle } from '../components/PageTitle'
import { SourcegraphContext } from '../jscontext'
import { PageRoutes } from '../routes.constants'
import { eventLogger } from '../tracking/eventLogger'

import { SourcegraphIcon } from './icons'
import { getReturnTo } from './SignInSignUpCommon'

import RequestAccessSignUpCommonStyles from './SignInSignUpCommon.module.scss'

interface RequestAccessPageProps {
    authenticatedUser: AuthenticatedUser | null
    context: Pick<
        SourcegraphContext,
        | 'allowSignup'
        | 'authProviders'
        | 'sourcegraphDotComMode'
        | 'xhrHeaders'
        | 'resetPasswordEnabled'
        | 'experimentalFeatures'
    >
    isSourcegraphDotCom: boolean
}

export const RequestAccessPage: React.FunctionComponent<React.PropsWithChildren<RequestAccessPageProps>> = props => {
    useEffect(() => eventLogger.logPageView('RequestAccessPage'), [])

    const location = useLocation()
    const [error, setError] = useState<Error | null>(null)
    const [loading, setLoading] = useState<boolean>(false)

    if (props.authenticatedUser) {
        const returnTo = getReturnTo(location)
        return <Navigate to={returnTo} replace={true} />
    }

    // TODO: check case when allowSignup=true and no seats
    // TODO: move this to a shared variable with the one in the SignInPage
    const showRequestAccess =
        !props.isSourcegraphDotCom &&
        !props.context.allowSignup &&
        props.context.experimentalFeatures?.requestAccess?.enabled !== false

    if (!showRequestAccess) {
        return <Navigate to={PageRoutes.SignIn} replace={true} />
    }

    const body = (
        <div className={classNames('mb-4 pb-5', RequestAccessSignUpCommonStyles.signinPageContainer)}>
            {error && <ErrorAlert className="mt-4 mb-0 text-left" error={error} />}
            <div
                className={classNames(
                    'test-RequestAccess-form rounded p-4 my-3',
                    RequestAccessSignUpCommonStyles.signinSignupForm,
                    error ? 'mt-3' : 'mt-4'
                )}
            >
                <Form onSubmit={noop}>
                    <Input
                        id="name"
                        label={<Text alignment="left">Name</Text>}
                        onChange={noop}
                        required={true}
                        // value={usernameOrEmail}
                        disabled={loading}
                        autoCapitalize="off"
                        autoFocus={true}
                        className="form-group"
                        placeholder="Your name"
                        autoComplete="name"
                    />
                    <Input
                        id="email"
                        label={<Text alignment="left">Email Address</Text>}
                        onChange={noop}
                        required={true}
                        // value={usernameOrEmail}
                        disabled={loading}
                        autoCapitalize="off"
                        autoFocus={true}
                        placeholder="Your work email to get access"
                        className="form-group"
                        autoComplete="email"
                    />
                    <TextArea
                        onChange={noop}
                        className="mb-4"
                        // value={value}
                        label="Extra information"
                        placeholder="Use this field to provide extra info for your request access"
                    />
                    <div className={classNames('form-group')}>
                        <Button display="block" type="submit" disabled={loading} variant="primary">
                            {loading ? <LoadingSpinner /> : 'Request access'}
                        </Button>
                    </div>
                </Form>
            </div>
            <Text className="mt-3">
                Already have an account? <Link to={`/sign-in${location.search}`}>Sign in</Link>
            </Text>
        </div>
    )

    return (
        <div className={RequestAccessSignUpCommonStyles.signinSignupPage}>
            <PageTitle title="Request access" />
            <HeroPage
                icon={SourcegraphIcon}
                iconLinkTo={props.context.sourcegraphDotComMode ? '/search' : undefined}
                iconClassName="bg-transparent"
                lessPadding={true}
                title="Request access to Sourcegraph"
                body={body}
            />
        </div>
    )
}
