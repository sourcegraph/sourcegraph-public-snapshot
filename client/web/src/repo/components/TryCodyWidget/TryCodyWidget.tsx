import React, { useCallback, useEffect } from 'react'

import { mdiClose } from '@mdi/js'
import classNames from 'classnames'

import { useTemporarySetting } from '@sourcegraph/shared/src/settings/temporary'
import type { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { useIsLightTheme } from '@sourcegraph/shared/src/theme'
import { Button, H2, H4, Icon, Link, Text } from '@sourcegraph/wildcard'

import type { AuthenticatedUser } from '../../../auth'
import { ExternalsAuth } from '../../../auth/components/ExternalsAuth'
import { MarketingBlock } from '../../../components/MarketingBlock'
import type { AuthProvider, SourcegraphContext } from '../../../jscontext'
import { EventName } from '../../../util/constants'

import { GlowingCodySVG, MeetCodySVG } from './WidgetIcons'

import styles from './TryCodyWidget.module.scss'

const AUTO_DISMISS_ON_EVENTS = new Set([EventName.CODY_SIDEBAR_CHAT_OPENED, EventName.CODY_CHAT_SUBMIT])

interface WidgetContentProps extends TelemetryProps {
    type: 'blob' | 'repo'
    theme?: 'light' | 'dark'
    isSourcegraphDotCom: boolean
}

interface NoAuhWidgetContentProps extends WidgetContentProps {
    context: Pick<SourcegraphContext, 'authProviders'>
}

function useTryCodyWidget(telemetryService: TelemetryProps['telemetryService']): {
    isDismissed: boolean | undefined
    onDismiss: () => void
} {
    // `isDismissed = true` maintain the initial concealment of the CTA when loading the settings
    const [isDismissed = true, setIsDismissed] = useTemporarySetting('cody.blobPageCta.dismissed', false)

    const onDismiss = useCallback(() => {
        setIsDismissed(true)
    }, [setIsDismissed])

    // Listen for telemetry events to auto dismiss the widget
    useEffect(() => {
        if (isDismissed) {
            return
        }

        return telemetryService.addEventLogListener?.(eventName => {
            if (AUTO_DISMISS_ON_EVENTS.has(eventName as EventName)) {
                onDismiss()
            }
        })
    }, [telemetryService, isDismissed, onDismiss])

    return { isDismissed, onDismiss }
}

const NoAuthWidgetContent: React.FC<NoAuhWidgetContentProps> = ({ type, telemetryService, context }) => {
    const logEvent = (provider: AuthProvider['serviceType']): void => {
        const eventType = provider === 'builtin' ? 'form' : provider
        const eventPage = type === 'blob' ? 'Blobview' : 'RepositoryPage'
        const eventArguments = {
            type: eventType,
            page: eventPage,
            description: '',
        }
        telemetryService.log(EventName.SIGNUP_INITIATED, eventArguments, eventArguments)
    }

    const title = type === 'blob' ? 'Sign up to get Cody, our AI assistant, free' : 'Meet Cody, your AI assistant'

    return (
        <>
            <MeetCodySVG />
            <div className="flex-grow-1">
                <H2 className={styles.cardTitle}>{title}</H2>
                <Text className={styles.cardDescription}>
                    Cody combines an LLM with the context of Sourcegraph's code graph on public code or your code at
                    work.{' '}
                </Text>
                <div className={styles.authButtonsWrap}>
                    <ExternalsAuth
                        context={context}
                        githubLabel="GitHub"
                        gitlabLabel="GitLab"
                        googleLabel="Google"
                        withCenteredText={true}
                        onClick={logEvent}
                        ctaClassName={styles.authButton}
                    />
                </div>
                <Text className="mb-2 mt-2">
                    By registering, you agree to our{' '}
                    <Link
                        to="https://sourcegraph.com/terms"
                        className={styles.termsLink}
                        target="_blank"
                        rel="noopener"
                    >
                        Terms of Service
                    </Link>{' '}
                    and{' '}
                    <Link
                        to="https://sourcegraph.com/terms/privacy"
                        className={styles.termsLink}
                        target="_blank"
                        rel="noopener"
                    >
                        Privacy Policy
                    </Link>
                </Text>
            </div>
        </>
    )
}

const AuthUserWidgetContent: React.FC<WidgetContentProps> = ({ type, theme, isSourcegraphDotCom }) => {
    const { title, useCases, image } = isSourcegraphDotCom
        ? type === 'blob'
            ? {
                  title: 'Try Cody on public code',
                  useCases: ['Select code in the file below', 'Select an action with Cody widget'],
                  image: `https://storage.googleapis.com/sourcegraph-assets/app-images/cody-action-bar-${theme}.png`,
              }
            : {
                  title: 'Try Cody on this repository',
                  useCases: [
                      'Click the Ask Cody button above and to the right of this banner',
                      'Ask Cody a question like “Explain the structure of this repository”',
                  ],
                  image: `https://storage.googleapis.com/sourcegraph-assets/app-images/cody-chat-banner-image-${theme}.png`,
              }
        : type === 'blob'
        ? {
              title: 'Try Cody on this file',
              useCases: ['Select code in the file below', 'Select an action with Cody widget'],
              image: `https://storage.googleapis.com/sourcegraph-assets/app-images/cody-action-bar-${theme}.png`,
          }
        : {
              title: 'Try Cody on this repository',
              useCases: [
                  'Click the Ask Cody button above and to the right of this banner',
                  'Ask Cody a question like “Explain the structure of this repository”',
              ],
              image: `https://storage.googleapis.com/sourcegraph-assets/app-images/cody-chat-banner-image-${theme}.png`,
          }

    return (
        <>
            <div className="d-flex pb-3">
                <GlowingCodySVG />
                <div className="d-flex flex-column flex-grow-1 justify-content-center flex-shrink-0">
                    <H4 as="h2" className={styles.cardTitle}>
                        {title}
                    </H4>
                    <ol className={classNames('m-0 pl-4', styles.cardList)}>
                        {useCases.map(useCase => (
                            <Text key={useCase} as="li">
                                {useCase}
                            </Text>
                        ))}
                    </ol>
                </div>
            </div>
            <div className={classNames('d-flex justify-content-center', styles.cardImages)}>
                <img src={image} alt="Cody" className={classNames(styles.cardImage, 'percy-hide')} />
            </div>
        </>
    )
}

interface TryCodyWidgetProps extends TelemetryProps {
    className?: string
    type: 'blob' | 'repo'
    authenticatedUser: AuthenticatedUser | null
    context: Pick<SourcegraphContext, 'authProviders'>
    isSourcegraphDotCom: boolean
}

export const TryCodyWidget: React.FC<TryCodyWidgetProps> = ({
    className,
    telemetryService,
    authenticatedUser,
    context,
    type,
    isSourcegraphDotCom,
}) => {
    const isLightTheme = useIsLightTheme()
    const { isDismissed, onDismiss } = useTryCodyWidget(telemetryService)
    useEffect(() => {
        if (isDismissed) {
            return
        }
        const eventPage = type === 'blob' ? 'BlobPage' : 'RepoPage'
        telemetryService.log(EventName.TRY_CODY_WEB_ONBOARDING_DISPLAYED, { type: eventPage }, { type: eventPage })
    }, [isDismissed, telemetryService, type])

    if (isDismissed) {
        return null
    }

    return (
        <MarketingBlock
            wrapperClassName={classNames(
                className,
                type === 'blob' ? styles.blobCardWrapper : styles.repoCardWrapper,
                'mb-2'
            )}
            contentClassName={classNames(
                'd-flex position-relative pb-0 overflow-auto justify-content-between',
                styles.card,
                !authenticatedUser && styles.noAuthCard
            )}
            variant="thin"
        >
            {authenticatedUser ? (
                <AuthUserWidgetContent
                    type={type}
                    theme={isLightTheme ? 'light' : 'dark'}
                    telemetryService={telemetryService}
                    isSourcegraphDotCom={isSourcegraphDotCom}
                />
            ) : (
                <NoAuthWidgetContent
                    telemetryService={telemetryService}
                    type={type}
                    context={context}
                    isSourcegraphDotCom={isSourcegraphDotCom}
                />
            )}
            <Button className={classNames(styles.closeButton, 'position-absolute mt-2')} onClick={onDismiss}>
                <Icon svgPath={mdiClose} aria-label="Close try Cody widget" />
            </Button>
        </MarketingBlock>
    )
}
