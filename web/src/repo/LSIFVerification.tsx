import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import CheckIcon from 'mdi-react/CheckIcon'
import * as React from 'react'
import { ErrorLike, isErrorLike, asError } from '../../../shared/src/util/errors'
import { CopyableText } from '../components/CopyableText'
import { Subscription, from } from 'rxjs'
import { catchError } from 'rxjs/operators'

interface Props {
    repoName: string
}

interface State {
    challengeOrError?: string | ErrorLike
    verifying?: boolean
    tokenOrError?: string | ErrorLike
}

interface ChallengeResponse {
    challenge: string
}

type VerifyResponse =
    | {
          failure: string
      }
    | { token: string }

async function fetchChallenge(): Promise<string> {
    const response = await fetch(new URL('/.api/lsif/challenge', window.location.href).href, {
        headers: {
            'X-Requested-With': 'Sourcegraph',
        },
    })
    if (response.status !== 200) {
        throw new Error(
            'Unable to fetch the LSIF challenge. Make sure lsifUploadSecret is set in the site configuration.'
        )
    }
    const json: ChallengeResponse = await response.json()
    return json.challenge
}

export class LSIFVerification extends React.PureComponent<Props, State> {
    private subscriptions = new Subscription()

    constructor(props: Props) {
        super(props)

        this.state = {}
    }

    public componentDidMount(): void {
        this.subscriptions.add(
            from(fetchChallenge())
                .pipe(catchError(error => [asError(error)]))
                .subscribe(challengeOrError => {
                    this.setState({ challengeOrError })
                })
        )
    }

    public render(): JSX.Element | null {
        if (isErrorLike(this.state.challengeOrError)) {
            return <div className="alert alert-danger">{this.state.challengeOrError.message}</div>
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

                        {this.state.challengeOrError ? (
                            <>
                                <CopyableText text={this.state.challengeOrError} size={16} />
                                <button
                                    type="button"
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
            if ('failure' in response) {
                throw new Error(response.failure)
            }
            this.setState({ tokenOrError: response.token })
        } catch (error) {
            this.setState({ tokenOrError: error })
        } finally {
            this.setState({ verifying: false })
        }
    }
}
