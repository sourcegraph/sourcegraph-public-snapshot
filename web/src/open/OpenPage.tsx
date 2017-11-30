import * as H from 'history'
import * as React from 'react'
import { PageTitle } from '../components/PageTitle'
import { eventLogger } from '../tracking/eventLogger'
import { basename, dirname } from '../util/path'

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
                <h2>Open in {validSchemes[this.state.scheme]}:</h2>
                <pre>
                    <code className="open-page__commands alert alert-primary">{this.commandsToRun()}</code>
                </pre>
                <p className="open-page__info">
                    Don't have the editor yet?{' '}
                    <a href="https://about.sourcegraph.com/beta/201708/#beta">Download Sourcegraph Editor</a>.
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

    private commandsToRun(): string {
        const params = new URLSearchParams(this.props.location.search)

        const commands: string[] = []

        const sanitize = (s: string): string => s.replace(/[^a-zA-Z0-9_/@:.-]/g, '_')

        const repo = params.get('repo')
        if (repo) {
            const repoParts: string[] = [basename(repo).replace(/\.git$/, '')]
            const repoDirname = dirname(repo)
            if (repoDirname && repoDirname !== '.') {
                repoParts.unshift(basename(repoDirname).replace(/^.*:/, ''))
            }
            commands.push(`cd /path/to/${sanitize(repoParts.join('/'))}`)
        }

        const revision = params.get('revision')
        if (revision) {
            commands.push(`git checkout ${sanitize(revision)}`)
        }

        const cmd = this.state.scheme
        let path = params.get('path')
        const selection = params.get('selection')
        if (path) {
            if (path.startsWith('--')) {
                path = `-- ${path}` // don't interpret -- in path as CLI flags
            }
            const args = selection ? `--goto ${sanitize(path)}:${sanitize(selection)}` : path
            commands.push(`${cmd} ${args}`)
        } else {
            commands.push(`${cmd} -n .`)
        }

        return commands.join('\n')
    }
}
