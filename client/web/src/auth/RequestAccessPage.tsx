import React, { useEffect, useState } from 'react'

import classNames from 'classnames'
import { Navigate, Route, Routes, useLocation, useNavigate } from 'react-router-dom-v5-compat'

import { Text, Link, ErrorAlert, Form, Input, Button, LoadingSpinner, TextArea, Label } from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../auth'
import { HeroPage } from '../components/HeroPage'
import { PageTitle } from '../components/PageTitle'
import { SourcegraphContext } from '../jscontext'
import { PageRoutes } from '../routes.constants'
import { eventLogger } from '../tracking/eventLogger'
import { checkIsRequestAccessEnabled } from '../util/checkIsRequestAccessEnabled'

import { SourcegraphIcon } from './icons'
import { getReturnTo } from './SignInSignUpCommon'

import RequestAccessSignUpCommonStyles from './SignInSignUpCommon.module.scss'

// TODO: add tests
interface RequestAccessFormProps {
    onSuccess: () => void
    onError: (error: any) => void
    context: Pick<SourcegraphContext, 'xhrHeaders'>
}
const RequestAccessForm: React.FunctionComponent<RequestAccessFormProps> = ({ onSuccess, onError, context }) => {
    const [loading, setLoading] = useState<boolean>(false)
    const [email, setEmail] = useState<string>('')
    const [name, setName] = useState<string>('')
    const [additionalInfo, setAdditionalInfo] = useState<string>('')

    const handleSubmit = async (event: React.FormEvent<HTMLFormElement>): Promise<void> => {
        event.preventDefault()
        if (loading) {
            return
        }
        setLoading(true)
        try {
            const response = await fetch('/-/request-access', {
                credentials: 'same-origin',
                method: 'POST',
                headers: {
                    ...context.xhrHeaders,
                    Accept: 'application/json',
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    email,
                    name,
                    additionalInfo,
                }),
            })

            if (!response.ok) {
                const text = await response.text()
                onError(new Error(response.statusText + ': ' + text))
            } else {
                onSuccess()
            }
        } catch (error) {
            onError(error)
        } finally {
            setLoading(false)
        }
    }
    return (
        <Form onSubmit={handleSubmit}>
            <Label className="w-100">
                <Text alignment="left" className="mb-2">
                    Name
                </Text>
                <Input
                    id="name"
                    onChange={(event: React.ChangeEvent<HTMLInputElement>) => setName(event.target.value)}
                    required={true}
                    value={name}
                    disabled={loading}
                    autoCapitalize="off"
                    autoFocus={true}
                    className="form-group"
                    placeholder="Your name"
                    autoComplete="name"
                />
            </Label>
            <Label className="w-100">
                <Text alignment="left" className="mb-2">
                    Email Address
                </Text>
                <Input
                    id="email"
                    onChange={(event: React.ChangeEvent<HTMLInputElement>) => setEmail(event.target.value)}
                    required={true}
                    value={email}
                    disabled={loading}
                    autoCapitalize="off"
                    autoFocus={true}
                    placeholder="Your work email to get access"
                    className="form-group"
                    autoComplete="email"
                />
            </Label>
            <Label className="w-100">
                <Text alignment="left" className="mb-2">
                    Extra information
                </Text>
                <TextArea
                    onChange={(event: React.ChangeEvent<HTMLTextAreaElement>) => setAdditionalInfo(event.target.value)}
                    className="mb-4"
                    value={additionalInfo}
                    placeholder="Use this field to provide extra info for your request access"
                />
            </Label>

            <div className={classNames('form-group')}>
                <Button display="block" type="submit" disabled={loading} variant="primary">
                    {loading ? <LoadingSpinner /> : 'Request access'}
                </Button>
            </div>
        </Form>
    )
}

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

export const RequestAccessPage: React.FunctionComponent<React.PropsWithChildren<RequestAccessPageProps>> = ({
    isSourcegraphDotCom,
    context,
    authenticatedUser,
}) => {
    useEffect(() => eventLogger.logPageView('RequestAccessPage'), [])
    const location = useLocation()
    const navigate = useNavigate()
    const [error, setError] = useState<Error | null>(null)
    const isRequestAccessEnabled = checkIsRequestAccessEnabled(
        isSourcegraphDotCom,
        context.allowSignup,
        context.experimentalFeatures.requestAccess
    )

    if (authenticatedUser) {
        const returnTo = getReturnTo(location)
        return <Navigate to={returnTo} replace={true} />
    }

    if (!isRequestAccessEnabled) {
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
                <Routes>
                    <Route
                        path="done"
                        element={
                            <Text>
                                Thank your for confirming your email [Name]. We notified the admin of this instance of
                                your request.
                            </Text>
                        }
                    />
                    <Route
                        path=""
                        element={
                            <RequestAccessForm
                                onError={setError}
                                context={context}
                                onSuccess={() => navigate('done')}
                            />
                        }
                    />
                </Routes>
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
                iconLinkTo={context.sourcegraphDotComMode ? '/search' : undefined}
                iconClassName="bg-transparent"
                lessPadding={true}
                title="Request access to Sourcegraph"
                body={body}
            />
        </div>
    )
}
