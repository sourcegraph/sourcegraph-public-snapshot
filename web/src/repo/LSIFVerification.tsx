import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import CheckIcon from 'mdi-react/CheckIcon'
import * as React from 'react'
import { ErrorLike, isErrorLike } from '../../../shared/src/util/errors'
import { CopyableText } from '../components/CopyableText'

interface Props {
    repoName: string
}

interface State {
    challenge?: string
    verifying?: boolean
    error?: string
    tokenOrError?: string | ErrorLike
}

interface ChallengeResponse {
    Challenge: string
}

type VerifyResponse =
    | {
          Failure: string
      }
    | { Token: string }

async function fetchChallenge(repoName: string): Promise<string> {
    const response: ChallengeResponse = await (await fetch(new URL('/.api/lsif/challenge', window.location.href).href, {
        headers: {
            'X-Requested-With': 'Sourcegraph',
        },
    })).json()
    return response.Challenge
}

export class LSIFVerification extends React.PureComponent<Props, State> {
    constructor(props: Props) {
        super(props)

        this.state = {}
    }

    public async componentDidMount(): Promise<void> {
        try {
            this.setState({
                challenge: await fetchChallenge(this.props.repoName),
            })
        } catch (error) {
            this.setState({ error })
        }
    }

    public render(): JSX.Element | null {
        if (this.state.error) {
            return <div className="alert alert-danger">{this.state.error.toString()}</div>
        }

        // Only verification for GitHub will been implemented for GopherCon.
        if (!this.props.repoName.startsWith('github.com')) {
            return <div>LSIF is currently only supported for GitHub.com repositories.</div>
        }

        return (
            <div className="lsif-verification">
                {this.state.tokenOrError && !isErrorLike(this.state.tokenOrError) ? (
                    <div className="alert alert-success">
                        <CheckIcon className="icon-inline" /> Topic found. Here's the LSIF upload token:
                        <CopyableText text={this.state.tokenOrError} size={128} />
                        You can remove the topic now.
                    </div>
                ) : (
                    <>
                        <div>
                            To get an LSIF upload token for this repository, prove that you own the repository by adding
                            a GitHub topic to <a href={`https://${this.props.repoName}`}>{this.props.repoName}</a> with
                            the following name:
                        </div>

                        {this.state.challenge ? (
                            <>
                                <CopyableText text={this.state.challenge} size={16} />
                                <button
                                    className="btn btn-primary"
                                    disabled={this.state.verifying}
                                    onClick={this.verify}
                                >
                                    {this.state.verifying && <LoadingSpinner className="icon-inline" />}
                                    Check now
                                </button>
                                {isErrorLike(this.state.tokenOrError) && (
                                    <div className="alert alert-danger">{this.state.tokenOrError.toString()}</div>
                                )}
                            </>
                        ) : (
                            <div>Loading challenge...</div>
                        )}
                    </>
                )}
            </div>
        )
    }

    private verify = async () => {
        this.setState({ tokenOrError: undefined, verifying: true })
        try {
            const url = new URL('/.api/lsif/verify', window.location.href)
            url.searchParams.set('repository', this.props.repoName)
            const response: VerifyResponse = await (await fetch(url.href, {
                headers: {
                    'X-Requested-With': 'Sourcegraph',
                },
            })).json()
            if ('Failure' in response) {
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
