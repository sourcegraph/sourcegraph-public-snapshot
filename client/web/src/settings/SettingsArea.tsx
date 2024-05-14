import * as React from 'react'

import classNames from 'classnames'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import { combineLatest, type Observable, Subject, Subscription } from 'rxjs'
import { catchError, distinctUntilChanged, map, startWith, switchMap } from 'rxjs/operators'

import { asError, createAggregateError, type ErrorLike, isErrorLike, logger } from '@sourcegraph/common'
import { gql } from '@sourcegraph/http-client'
import type { Scalars } from '@sourcegraph/shared/src/graphql-operations'
import type { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import type { SettingsCascadeProps, SettingsSubject } from '@sourcegraph/shared/src/settings/settings'
import type { TelemetryV2Props } from '@sourcegraph/shared/src/telemetry'
import type { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { EVENT_LOGGER } from '@sourcegraph/shared/src/telemetry/web/eventLogger'
import { LoadingSpinner, PageHeader, ErrorMessage } from '@sourcegraph/wildcard'

import settingsSchemaJSON from '../../../../schema/settings.schema.json'
import type { AuthenticatedUser } from '../auth'
import { queryGraphQL } from '../backend/graphql'
import { HeroPage } from '../components/HeroPage'
import type { SettingsCascadeResult } from '../graphql-operations'

import { SettingsPage } from './SettingsPage'

/** Props shared by SettingsArea and its sub-pages. */
interface SettingsAreaPageCommonProps extends PlatformContextProps, SettingsCascadeProps, TelemetryProps {
    /** The subject whose settings to edit. */
    subject: Pick<SettingsSubject, '__typename' | 'id'>

    /**
     * The currently authenticated user, NOT (necessarily) the user who is the subject of the page.
     */
    authenticatedUser: AuthenticatedUser | null

    isLightTheme: boolean
}

interface SettingsData {
    subjects: SettingsSubject[]
}

/** Properties passed to all pages in the settings area. */
export interface SettingsAreaPageProps extends SettingsAreaPageCommonProps, TelemetryV2Props {
    /** The settings data, or null if the subject has no settings yet. */
    data: SettingsData

    /** Called when the page updates the subject's settings. */
    onUpdate: () => void
}

interface Props extends SettingsAreaPageCommonProps {
    className?: string
    extraHeader?: JSX.Element
}

const LOADING = 'loading' as const

interface State {
    /**
     * The data, loading, or an error.
     */
    dataOrError: typeof LOADING | SettingsData | ErrorLike
}

interface SettingsSubjects {
    subjects: SettingsSubject[]
}

/**
 * A settings area with a top-level JSON editor and sub-pages for editing nested settings values.
 */
export class SettingsArea extends React.Component<Props, State> {
    public state: State = { dataOrError: LOADING }

    private componentUpdates = new Subject<Props>()
    private refreshRequests = new Subject<void>()
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        EVENT_LOGGER.logViewEvent(`Settings${this.props.subject.__typename}`)
        switch (this.props.subject.__typename) {
            case 'User': {
                this.props.platformContext.telemetryRecorder.recordEvent('user.settings', 'view')
                break
            }
            case 'Org': {
                this.props.platformContext.telemetryRecorder.recordEvent('org.settings', 'view')
                break
            }
            case 'Site': {
                this.props.platformContext.telemetryRecorder.recordEvent('admin.settings', 'view')
                break
            }
            case 'DefaultSettings': {
                this.props.platformContext.telemetryRecorder.recordEvent('defaultSettings', 'view')
                break
            }
            case 'Client': {
                this.props.platformContext.telemetryRecorder.recordEvent('client.settings', 'view')
            }
            default: {
                this.props.platformContext.telemetryRecorder.recordEvent('otherSettings', 'view')
                break
            }
        }

        // Load settings.
        this.subscriptions.add(
            combineLatest([
                this.componentUpdates.pipe(
                    map(props => props.subject),
                    distinctUntilChanged()
                ),
                this.refreshRequests.pipe(startWith<void>(undefined)),
            ])
                .pipe(
                    switchMap(([{ id }]) =>
                        fetchSettingsCascade(id).pipe(
                            map(cascade => ({ subjects: cascade.subjects, settingsJSONSchema: settingsSchemaJSON })),
                            catchError(error => [asError(error)]),
                            map(dataOrError => ({ dataOrError }))
                        )
                    )
                )
                .subscribe(
                    stateUpdate => this.setState(stateUpdate),
                    error => logger.error(error)
                )
        )

        this.componentUpdates.next(this.props)
    }

    public componentDidUpdate(): void {
        this.componentUpdates.next(this.props)
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        if (this.state.dataOrError === LOADING) {
            return <LoadingSpinner inline={false} />
        }
        if (isErrorLike(this.state.dataOrError)) {
            return (
                <HeroPage
                    icon={AlertCircleIcon}
                    title="Error"
                    subtitle={<ErrorMessage error={this.state.dataOrError} />}
                />
            )
        }

        let term: string
        switch (this.props.subject.__typename) {
            case 'User': {
                term = 'User'
                break
            }
            case 'Org': {
                term = 'Organization'
                break
            }
            case 'Site': {
                term = 'Global'
                break
            }
            case 'DefaultSettings': {
                term = 'Default settings'
                break
            }
            default: {
                term = 'Unknown'
                break
            }
        }

        const transferProps: SettingsAreaPageProps = {
            data: this.state.dataOrError,
            subject: this.props.subject,
            authenticatedUser: this.props.authenticatedUser,
            onUpdate: this.onUpdate,
            isLightTheme: this.props.isLightTheme,
            platformContext: this.props.platformContext,
            settingsCascade: this.props.settingsCascade,
            telemetryService: this.props.telemetryService,
            telemetryRecorder: this.props.platformContext.telemetryRecorder,
        }

        return (
            <div className={classNames('h-100 d-flex flex-column', this.props.className)}>
                <PageHeader className="mb-3">
                    <PageHeader.Heading as="h3" styleAs="h2">
                        <PageHeader.Breadcrumb>{`${term} settings`}</PageHeader.Breadcrumb>
                    </PageHeader.Heading>
                </PageHeader>
                {this.props.extraHeader}
                <SettingsPage {...transferProps} />
            </div>
        )
    }

    private onUpdate = (): void => this.refreshRequests.next()
}

function fetchSettingsCascade(subject: Scalars['ID']): Observable<SettingsSubjects> {
    return queryGraphQL<SettingsCascadeResult>(
        gql`
            query SettingsCascade($subject: ID!) {
                settingsSubject(id: $subject) {
                    settingsCascade {
                        subjects {
                            latestSettings {
                                id
                                contents
                            }
                        }
                    }
                }
            }
        `,
        { subject }
    ).pipe(
        map(({ data, errors }) => {
            if (!data?.settingsSubject) {
                throw createAggregateError(errors)
            }
            return data.settingsSubject.settingsCascade as SettingsSubjects
        })
    )
}
