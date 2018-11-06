import * as React from 'react'
import { Observable, of, Subject, Subscription } from 'rxjs'
import { catchError, distinctUntilChanged, filter, map, mapTo, share, switchMap, tap } from 'rxjs/operators'
import { getExtensionVersionSync } from '../../browser/runtime'
import { AccessToken, FeatureFlags } from '../../browser/types'
import { ERAUTHREQUIRED, ErrorLike, isErrorLike } from '../../shared/backend/errors'
import { propertyIsDefined } from '../../shared/util/types'
import { GQL } from '../../types/gqlschema'
import { OptionsMenu, OptionsMenuProps } from './Menu'
import { ConnectionErrors } from './ServerURLForm'

export interface OptionsContainerProps {
    sourcegraphURL: string

    ensureValidSite: (url: string) => Observable<any>
    fetchCurrentUser: (useToken: boolean) => Observable<GQL.IUser | undefined>

    setSourcegraphURL: (url: string) => void
    getConfigurableSettings: () => Observable<Partial<FeatureFlags>>
    setConfigurableSettings: (settings: Partial<FeatureFlags>) => Observable<Partial<FeatureFlags>>

    createAccessToken: (url: string) => Observable<AccessToken>
    getAccessToken: (url: string) => Observable<AccessToken | undefined>
    setAccessToken: (url: string, token: AccessToken) => void
    fetchAccessTokenIDs: (url: string) => Observable<Pick<AccessToken, 'id'>[]>
}

interface OptionsContainerState
    extends Pick<
            OptionsMenuProps,
            'isSettingsOpen' | 'status' | 'sourcegraphURL' | 'settings' | 'settingsHaveChanged' | 'connectionError'
        > {}

export class OptionsContainer extends React.Component<OptionsContainerProps, OptionsContainerState> {
    private version = getExtensionVersionSync()

    private urlUpdates = new Subject<string>()
    private settingsSaves = new Subject<any>()

    private subscriptions = new Subscription()

    constructor(props: OptionsContainerProps) {
        super(props)

        this.state = {
            status: 'connecting',
            sourcegraphURL: props.sourcegraphURL,
            isSettingsOpen: false,
            settingsHaveChanged: false,
            settings: {},
            connectionError: undefined,
        }

        const fetchingSite: Observable<string | ErrorLike> = this.urlUpdates.pipe(
            distinctUntilChanged(),
            filter(maybeURL => {
                let validURL = false
                try {
                    validURL = !!new URL(maybeURL)
                } catch (e) {
                    validURL = false
                }

                return validURL
            }),
            switchMap(url => {
                this.setState({ status: 'connecting', connectionError: undefined })
                return this.props.ensureValidSite(url).pipe(
                    map(() => url),
                    catchError(err => of(err))
                )
            }),
            catchError(err => of(err)),
            share()
        )

        this.subscriptions.add(
            fetchingSite.subscribe(res => {
                let url = ''

                if (isErrorLike(res)) {
                    this.setState({
                        status: 'error',
                        connectionError:
                            res.code === ERAUTHREQUIRED ? ConnectionErrors.AuthError : ConnectionErrors.UnableToConnect,
                    })
                    url = this.state.sourcegraphURL
                } else {
                    this.setState({ status: 'connected' })
                    url = res
                }

                props.setSourcegraphURL(url)
            })
        )

        this.subscriptions.add(
            // Ensure the site is valid.
            fetchingSite
                .pipe(
                    filter(urlOrError => !isErrorLike(urlOrError)),
                    map(urlOrError => urlOrError as string),
                    // Get the access token for this server if we have it.
                    switchMap(url => this.props.getAccessToken(url).pipe(map(token => ({ token, url })))),
                    switchMap(({ url, token }) =>
                        this.props.fetchCurrentUser(false).pipe(map(user => ({ user, token, url })))
                    ),
                    filter(propertyIsDefined('user')),
                    // Get the IDs for all access tokens for the user.
                    switchMap(({ token, user, url }) =>
                        this.props
                            .fetchAccessTokenIDs(user.id)
                            .pipe(map(usersTokenIDs => ({ usersTokenIDs, user, token, url })))
                    ),
                    // Make sure the token still exists on the server. If it
                    // does exits, use it, otherwise create a new one.
                    switchMap(({ user, token, usersTokenIDs, url }) => {
                        const tokenExists = token && usersTokenIDs.map(({ id }) => id).includes(token.id)

                        return token && tokenExists
                            ? of(undefined)
                            : this.props.createAccessToken(user.id).pipe(
                                  tap(createdToken => {
                                      this.props.setAccessToken(url, createdToken)
                                  }),
                                  mapTo(undefined)
                              )
                    })
                )
                .subscribe(() => {
                    // Don't do anything here, we already saved new tokens above.
                })
        )

        this.subscriptions.add(
            this.settingsSaves
                .pipe(switchMap(settings => props.setConfigurableSettings(settings)))
                .subscribe(settings => {
                    this.setState({
                        settings,
                        settingsHaveChanged: false,
                    })
                })
        )
    }

    public componentDidMount(): void {
        this.props.getConfigurableSettings().subscribe(settings => {
            this.setState({ settings })
        })

        this.urlUpdates.next(this.state.sourcegraphURL)
    }
    public componentDidUpdate(): void {
        this.urlUpdates.next(this.props.sourcegraphURL)
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): React.ReactNode {
        return (
            <OptionsMenu
                {...this.state}
                version={this.version}
                onURLChange={this.handleURLChange}
                onURLSubmit={this.handleURLSubmit}
                onSettingsClick={this.handleSettingsClick}
                onSettingsChange={this.handleSettingsChange}
                onSettingsSave={this.handleSettingsSave}
            />
        )
    }

    private handleURLChange = (value: string) => {
        this.setState({ sourcegraphURL: value })
    }

    private handleURLSubmit = () => {
        this.props.setSourcegraphURL(this.state.sourcegraphURL)
    }

    private handleSettingsClick = () => {
        this.setState(({ isSettingsOpen }) => ({ isSettingsOpen: !isSettingsOpen }))
    }

    private handleSettingsChange = (settings: any) => {
        this.setState({ settings, settingsHaveChanged: true })
    }

    private handleSettingsSave = () => {
        this.settingsSaves.next(this.state.settings)
    }
}
