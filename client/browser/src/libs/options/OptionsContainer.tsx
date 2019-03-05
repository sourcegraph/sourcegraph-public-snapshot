import * as React from 'react'
import { Observable, of, Subject, Subscription } from 'rxjs'
import { catchError, distinctUntilChanged, filter, map, share, switchMap } from 'rxjs/operators'
import * as GQL from '../../../../../shared/src/graphql/schema'
import * as permissions from '../../browser/permissions'
import { getExtensionVersionSync } from '../../browser/runtime'
import { ERAUTHREQUIRED, ErrorLike, isErrorLike } from '../../shared/backend/errors'
import { OptionsMenu, OptionsMenuProps } from './Menu'
import { ConnectionErrors } from './ServerURLForm'

export interface OptionsContainerProps {
    sourcegraphURL: string

    ensureValidSite: (url: string) => Observable<any>
    fetchCurrentUser: (useToken: boolean) => Observable<GQL.IUser | undefined>

    setSourcegraphURL: (url: string) => void

    toggleFeatureFlag: (key: string) => void
    featureFlags: { key: string; value: boolean }[]
}

interface OptionsContainerState
    extends Pick<
        OptionsMenuProps,
        'status' | 'sourcegraphURL' | 'connectionError' | 'isSettingsOpen' | 'urlHasPermissions'
    > {}

export class OptionsContainer extends React.Component<OptionsContainerProps, OptionsContainerState> {
    private version = getExtensionVersionSync()

    private urlUpdates = new Subject<string>()

    private subscriptions = new Subscription()

    constructor(props: OptionsContainerProps) {
        super(props)

        this.state = {
            status: 'connecting',
            sourcegraphURL: props.sourcegraphURL,
            urlHasPermissions: false,
            connectionError: undefined,
            isSettingsOpen: false,
        }

        const fetchingSite: Observable<string | ErrorLike> = this.urlUpdates.pipe(
            distinctUntilChanged(),
            map(url => url.replace(/\/$/, '')),
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
            fetchingSite.subscribe(async res => {
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

                const urlHasPermissions = await permissions.contains(url)
                this.setState({ urlHasPermissions })

                props.setSourcegraphURL(url)
            })
        )
    }

    public componentDidMount(): void {
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
                toggleFeatureFlag={this.props.toggleFeatureFlag}
                featureFlags={this.props.featureFlags}
                onSettingsClick={this.handleSettingsClick}
                requestPermissions={this.requestPermissions}
            />
        )
    }

    private requestPermissions = async () => {
        await permissions.request([this.state.sourcegraphURL])
    }

    private handleURLChange = (value: string) => {
        this.setState({ sourcegraphURL: value })
    }

    private handleURLSubmit = () => {
        this.props.setSourcegraphURL(this.state.sourcegraphURL)
    }

    private handleSettingsClick = () => {
        this.setState(state => ({
            isSettingsOpen: !state.isSettingsOpen,
        }))
    }
}
