import React, { FunctionComponent, useState } from 'react'
import { useLocation } from 'react-router'

import { LinkOrSpan } from '@sourcegraph/shared/src/components/LinkOrSpan'
import { TelemetryService } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { BrandLogo } from '@sourcegraph/web/src/components/branding/BrandLogo'
import { HeroPage } from '@sourcegraph/web/src/components/HeroPage'
import { Steps, Step, StepList, StepPanels, StepPanel, StepActions } from '@sourcegraph/wildcard/src/components/Steps'

import { PageTitle } from '../components/PageTitle'
import { UserAreaUserFields } from '../graphql-operations'
import { SourcegraphContext } from '../jscontext'
import { SelectAffiliatedRepos } from '../user/settings/repositories/SelectAffiliatedRepos'

import { getReturnTo } from './SignInSignUpCommon'
import { useExternalServices } from './useExternalServices'
import { CodeHostsConnection } from './welcome/CodeHostsConnection'
import { Footer } from './welcome/Footer'
import { StartSearching } from './welcome/StartSearching'

interface PostSignUpPage {
    authenticatedUser: UserAreaUserFields
    context: Pick<SourcegraphContext, 'authProviders' | 'experimentalFeatures' | 'sourcegraphDotComMode'>
    telemetryService: TelemetryService
}

interface Step {
    content: React.ReactElement
    isComplete: () => boolean
    prefetch?: () => void
    onNextButtonClick?: () => Promise<void>
}

export type RepoSelectionMode = 'all' | 'selected' | undefined

export const PostSignUpPage: FunctionComponent<PostSignUpPage> = ({
    authenticatedUser: user,
    context,
    telemetryService,
}) => {
    const location = useLocation()
    const skipPostSignup = (): void => history.push(getReturnTo(location))

    const [repoSelectionMode, setRepoSelectionMode] = useState<RepoSelectionMode>()

    const { externalServices, loadingServices, errorServices, refetchExternalServices } = useExternalServices(user.id)

    return (
        <>
            <LinkOrSpan to={getReturnTo(location)} className="post-signup-page__logo-link">
                <BrandLogo
                    className="position-absolute ml-3 mt-3 post-signup-page__logo"
                    isLightTheme={true}
                    variant="symbol"
                    onClick={skipPostSignup}
                />
            </LinkOrSpan>

            <div className="signin-signup-page post-signup-page">
                <PageTitle title="Welcome" />

                <HeroPage
                    lessPadding={true}
                    className="text-left"
                    body={
                        <div className="post-signup-page__container">
                            <h2>Get started with Sourcegraph</h2>
                            <p className="text-muted pb-3">
                                Three quick steps to add your repositories and get searching with Sourcegraph
                            </p>
                            <div className="mt-4 pb-3">
                                <Steps initialStep={1}>
                                    <StepList numeric={true}>
                                        <Step borderColor="purple">Connect with code hosts</Step>
                                        <Step borderColor="blue">Add repositories</Step>
                                        <Step borderColor="orange">Start searching</Step>
                                    </StepList>
                                    <StepPanels>
                                        <StepPanel>
                                            {externalServices && (
                                                <CodeHostsConnection
                                                    loading={loadingServices}
                                                    user={user}
                                                    error={errorServices}
                                                    externalServices={externalServices}
                                                    context={context}
                                                    refetch={refetchExternalServices}
                                                />
                                            )}
                                        </StepPanel>
                                        <StepPanel>
                                            <>
                                                <h3>Add repositories</h3>
                                                <p className="text-muted">
                                                    Choose repositories you own or collaborate on from your code hosts
                                                    to search with Sourcegraph. We’ll sync and index these repositories
                                                    so you can search your code all in one place.
                                                </p>
                                                <SelectAffiliatedRepos
                                                    authenticatedUser={user}
                                                    onRepoSelectionModeChange={setRepoSelectionMode}
                                                    telemetryService={telemetryService}
                                                />
                                            </>
                                        </StepPanel>
                                        <StepPanel>
                                            <StartSearching user={user} repoSelectionMode={repoSelectionMode} />
                                        </StepPanel>
                                    </StepPanels>
                                    <StepActions>
                                        <Footer />
                                    </StepActions>
                                </Steps>
                            </div>
                        </div>
                    }
                />
            </div>
        </>
    )
}
