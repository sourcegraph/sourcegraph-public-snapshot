import * as React from 'react'
import { Observable, of, Subject, Subscription } from 'rxjs'
import { catchError, distinctUntilChanged, map, share, switchMap, concatMap } from 'rxjs/operators'
import { ErrorLike, isErrorLike } from '../../../../shared/src/util/errors'
import { getExtensionVersion, InvalidSourcegraphURLError } from '../../shared/util/context'
import { OptionsMenu, OptionsMenuProps } from './OptionsMenu'

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
        'statusOrError' | 'sourcegraphURL' | 'isSettingsOpen' | 'isActivated' | 'urlHasPermissions' | 'currentTabStatus'
    > {}

export class OptionsContainer extends React.Component<OptionsContainerProps, OptionsContainerState> {
    private version = getExtensionVersion()

    private urlUpdates = new Subject<string>()

    private activationClicks = new Subject<boolean>()

    private subscriptions = new Subscription()

    constructor(props: OptionsContainerProps) {
        super(props)

        this.state = {
            statusOrError: 'connecting',
            sourcegraphURL: props.sourcegraphURL,
            isActivated: props.isActivated,
            urlHasPermissions: false,
            isSettingsOpen: false,
        }

        const fetchingSite: Observable<string | ErrorLike> = this.urlUpdates.pipe(
            distinctUntilChanged(),
            switchMap(url => {
                try {
                    new URL(url)
                } catch (err) {
                    throw new InvalidSourcegraphURLError(url)
                }
                this.setState({ statusOrError: 'connecting' })
                return this.props.ensureValidSite(url).pipe(
                    map(() => url),
                    catchError(err => of(err))
                )
            }),
            catchError(err => of(err)),
            share()
        )

        this.subscriptions.add(
            // eslint-disable-next-line @typescript-eslint/no-misused-promises
            fetchingSite.subscribe(async res => {
                let url = ''

                if (isErrorLike(res)) {
                    this.setState({
                        statusOrError: res,
                    })
                    url = this.state.sourcegraphURL
                } else {
                    this.setState({ statusOrError: 'connected' })
                    url = res
                }

                const urlHasPermissions = await props.hasPermissions(url)
                this.setState({ urlHasPermissions })

                await props.setSourcegraphURL(url)
            })
        )

        props
            .fetchCurrentTabStatus()
            .then(currentTabStatus => this.setState(state => ({ ...state, currentTabStatus })))
            .catch(err => {
                console.log('Error fetching current tab status', err)
            })
    }

    public componentDidMount(): void {
        this.urlUpdates.next(this.state.sourcegraphURL)
        this.subscriptions.add(
            this.activationClicks
                .pipe(concatMap(isActivated => this.props.toggleExtensionDisabled(isActivated)))
                .subscribe()
        )
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
                isActivated={this.props.isActivated}
                toggleFeatureFlag={this.props.toggleFeatureFlag}
                featureFlags={this.props.featureFlags}
                onSettingsClick={this.handleSettingsClick}
                onToggleActivationClick={this.handleToggleActivationClick}
                requestPermissions={this.props.requestPermissions}
            />
        )
    }

    private handleURLChange = (value: string): void => {
        this.setState({ sourcegraphURL: value })
    }

    private handleURLSubmit = async (): Promise<void> => {
        await this.props.setSourcegraphURL(this.state.sourcegraphURL)
    }

    private handleSettingsClick = (): void => {
        this.setState(state => ({
            isSettingsOpen: !state.isSettingsOpen,
        }))
    }

    private handleToggleActivationClick = (value: boolean): void => this.activationClicks.next(value)
}
