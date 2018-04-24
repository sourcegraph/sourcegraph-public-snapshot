import * as H from 'history'
import * as React from 'react'
import { PageTitle } from '../components/PageTitle'
import { eventLogger } from '../tracking/eventLogger'

interface Props {
    location: H.Location
    history: H.History
}

interface State {
    /**
     * The scheme to use for opening native editor links.
     */
    scheme: string
}

const validSchemes: { [key: string]: string } = {
    src: 'Sourcegraph Editor',
    'src-insiders': 'Sourcegraph Insiders',
    'src-oss': 'Sourcegraph OSS',
}

const localStorageKey = 'open-native-schema'

/**
 * The open in editor page
 */
export class OpenPage extends React.Component<Props, State> {
    constructor(props: Props) {
        super(props)

        // Validate scheme from local storage.
        let scheme = window.localStorage.getItem(localStorageKey)
        if (!scheme || !validSchemes[scheme]) {
            scheme = 'src-insiders'
        }

        this.state = {
            scheme,
        }
    }

    public componentDidMount(): void {
        eventLogger.logViewEvent('Open')
    }

    public render(): JSX.Element | null {
        return (
            <div className="open-page">
                <PageTitle title={'Open in Sourcegraph - Sourcegraph'} />
                <h1>Opening in {validSchemes[this.state.scheme]}</h1>
                <p>
                    Your should be redirected in a few seconds.
                    <br />
                    Using a different build?{' '}
                    {Object.entries(validSchemes)
                        .filter(([scheme, name]) => scheme !== this.state.scheme)
                        .map(([scheme, name], i) => (
                            <span key={name}>
                                {i === 0 ? '' : ', '}
                                <a
                                    href=""
                                    // tslint:disable-next-line:jsx-no-lambda
                                    onClick={e => this.onClickLink(e, scheme)}
                                >
                                    {name}
                                </a>
                            </span>
                        ))}
                </p>
                <iframe className="open-page__iframe" src={`${this.state.scheme}:open?${this.props.location.search}`} />
            </div>
        )
    }

    private onClickLink = (e: React.MouseEvent<HTMLAnchorElement>, scheme: string): void => {
        e.preventDefault()

        // Event logging.
        const name = validSchemes[scheme]
        eventLogger.log('OpenDifferentEditorBuildClicked', {
            editorBuildName: name,
        })

        // Update local storage and component state.
        window.localStorage.setItem(localStorageKey, scheme)
        this.setState({ scheme })
    }
}
