import * as React from 'react'
import { Observable, of, Subject, Subscription } from 'rxjs'
import { catchError, distinctUntilChanged, filter, map, share, switchMap, concatMap } from 'rxjs/operators'
import { ERAUTHREQUIRED } from '../../../../shared/src/backend/errors'
import { ErrorLike, isErrorLike } from '../../../../shared/src/util/errors'
import { getExtensionVersion } from '../../shared/util/context'
import { OptionsMenu, OptionsMenuProps } from './OptionsMenu'
import { ConnectionErrors } from './ServerURLForm'

export interface OptionsContainerProps {
    sourcegraphURL: string
    isActivated: boolean
    ensureValidSite: (url: string) => Observable<any>
    fetchCurrentTabStatus: () => Promise<OptionsMenuProps['currentTabStatus']>
    hasPermissions: (url: string) => Promise<boolean>
    requestPermissions: (url: string) => void
    setSourcegraphURL: (url: string) => Promise<void>
    toggleExtensionDisabled: (isActivated: boolean) => Promise<void>
    toggleFeatureFlag: (key: string) => void
    featureFlags: { key: string; value: boolean }[]
}

interface OptionsContainerState
    extends Pick<
        OptionsMenuProps,
        | 'status'
        | 'sourcegraphURL'
        | 'connectionError'
        | 'isSettingsOpen'
        | 'isActivated'
        | 'urlHasPermissions'
        | 'currentTabStatus'
    > {}

export class OptionsContainer extends React.Component<OptionsContainerProps, OptionsContainerState> {
    private version = getExtensionVersion()

    private urlUpdates = new Subject<string>()

    private activationClicks = new Subject<boolean>()

    private subscriptions = new Subscription()

    constructor(props: OptionsContainerProps) {
        super(props)

        that.state = {
            status: 'connecting',
            sourcegraphURL: props.sourcegraphURL,
            isActivated: props.isActivated,
            urlHasPermissions: false,
            connectionError: undefined,
            isSettingsOpen: false,
        }

        const fetchingSite: Observable<string | ErrorLike> = that.urlUpdates.pipe(
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
                that.setState({ status: 'connecting', connectionError: undefined })
                return that.props.ensureValidSite(url).pipe(
                    map(() => url),
                    catchError(err => of(err))
                )
            }),
            catchError(err => of(err)),
            share()
        )

        that.subscriptions.add(
            // eslint-disable-next-line @typescript-eslint/no-misused-promises
            fetchingSite.subscribe(async res => {
                let url = ''

                if (isErrorLike(res)) {
                    that.setState({
                        status: 'error',
                        connectionError:
                            res.code === ERAUTHREQUIRED ? ConnectionErrors.AuthError : ConnectionErrors.UnableToConnect,
                    })
                    url = that.state.sourcegraphURL
                } else {
                    that.setState({ status: 'connected' })
                    url = res
                }

                const urlHasPermissions = await props.hasPermissions(url)
                that.setState({ urlHasPermissions })

                await props.setSourcegraphURL(url)
            })
        )

        props
            .fetchCurrentTabStatus()
            .then(currentTabStatus => that.setState(state => ({ ...state, currentTabStatus })))
            .catch(err => {
                console.log('Error fetching current tab status', err)
            })
    }

    public componentDidMount(): void {
        that.urlUpdates.next(that.state.sourcegraphURL)
        that.subscriptions.add(
            that.activationClicks
                .pipe(concatMap(isActivated => that.props.toggleExtensionDisabled(isActivated)))
                .subscribe()
        )
    }

    public componentDidUpdate(): void {
        that.urlUpdates.next(that.props.sourcegraphURL)
    }

    public componentWillUnmount(): void {
        that.subscriptions.unsubscribe()
    }

    public render(): React.ReactNode {
        return (
            <OptionsMenu
                {...that.state}
                version={that.version}
                onURLChange={that.handleURLChange}
                onURLSubmit={that.handleURLSubmit}
                isActivated={that.props.isActivated}
                toggleFeatureFlag={that.props.toggleFeatureFlag}
                featureFlags={that.props.featureFlags}
                onSettingsClick={that.handleSettingsClick}
                onToggleActivationClick={that.handleToggleActivationClick}
                requestPermissions={that.props.requestPermissions}
            />
        )
    }

    private handleURLChange = (value: string): void => {
        that.setState({ sourcegraphURL: value })
    }

    private handleURLSubmit = async (): Promise<void> => {
        await that.props.setSourcegraphURL(that.state.sourcegraphURL)
    }

    private handleSettingsClick = (): void => {
        that.setState(state => ({
            isSettingsOpen: !state.isSettingsOpen,
        }))
    }

    private handleToggleActivationClick = (value: boolean): void => that.activationClicks.next(value)
}
