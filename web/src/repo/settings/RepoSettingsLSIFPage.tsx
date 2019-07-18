import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import * as React from 'react'
import { ErrorLike, isErrorLike } from '../../../../shared/src/util/errors'
import { CopyableText } from '../../components/CopyableText'
import { eventLogger } from '../../tracking/eventLogger'
import { fetchRepositoryServiceType } from '../backend'

interface Props {
    repoName: string
}

interface State {
    serviceType?: string
    challenge?: string
    verifying: boolean
    error?: string
    tokenOrError?: string | ErrorLike
}

export class RepoSettingsLSIFPage extends React.PureComponent<Props, State> {
    constructor(props: Props) {
        super(props)

        this.state = {
            verifying: false,
        }
    }

    public componentDidMount(): void {
        eventLogger.logViewEvent('RepoSettingsLSIF')

        const url = new URL('/.api/lsif/challenge', window.location.href)
        url.searchParams.set('repository', this.props.repoName)
        ;(async () => {
            this.setState({
                serviceType: await fetchRepositoryServiceType(this.props.repoName).toPromise(),
                challenge: (await (await fetch(url.href, {
                    headers: {
                        'X-Requested-With': 'Sourcegraph',
                    },
                })).json()).Challenge,
            })
        })().catch(error => this.setState({ error }))
    }

    public render(): JSX.Element | null {
        if (this.state.error) {
            return <div>{this.state.error.toString()}</div>
        }

        // only verification for GitHub will been implemented for GopherCon
        if (this.state.serviceType !== 'github') {
            return <div>LSIF is not supported for {this.state.serviceType} repositories.</div>
        }

        return (
            <>
                <div>
                    To get an LSIF upload token for this repository, prove that you own the repository by adding a
                    GitHub topic with the following name:
                </div>
                {this.state.challenge ? (
                    <>
                        <CopyableText text={this.state.challenge} size={16} />
                        <button className="btn btn-primary" disabled={this.state.verifying} onClick={this.verify}>
                            {this.state.verifying && <LoadingSpinner className="icon-inline" />}
                            Verify
                        </button>
                        {this.state.tokenOrError !== undefined &&
                            (isErrorLike(this.state.tokenOrError) ? (
                                <div>{this.state.tokenOrError.toString()}</div>
                            ) : (
                                <CopyableText text={this.state.tokenOrError} size={128} />
                            ))}
                    </>
                ) : (
                    <div>Loading challenge...</div>
                )}
            </>
        )
    }

    private verify = async () => {
        this.setState({ tokenOrError: undefined, verifying: true })
        try {
            const url = new URL('/.api/lsif/verify', window.location.href)
            url.searchParams.set('repository', this.props.repoName)
            const response = await await (await fetch(url.href, {
                headers: {
                    'X-Requested-With': 'Sourcegraph',
                },
            })).json()
            if (response.Failure) {
                throw new Error(response.Failure)
            }
            this.setState({ tokenOrError: response.Token })
        } catch (error) {
            this.setState({ tokenOrError: error })
        } finally {
            this.setState({ verifying: false })
        }
    }
}
