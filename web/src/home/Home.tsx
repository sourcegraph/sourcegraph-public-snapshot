import CloseIcon from '@sourcegraph/icons/lib/Close'
import HelpIcon from '@sourcegraph/icons/lib/Help'
import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { SearchBox } from 'sourcegraph/search/SearchBox'
import { sourcegraphContext } from 'sourcegraph/util/sourcegraphContext'

interface Props extends RouteComponentProps<void> {}

interface State {
    helpVisible: boolean
}

const shortcutMofidier = navigator.platform.startsWith('Mac') ? 'ctrl' : 'cmd'

/**
 * The landing page
 */
export class Home extends React.Component<Props, State> {
    constructor(props: Props) {
        super(props)
        this.state = {
            helpVisible: localStorage.getItem('show-search-help') !== 'false'
        }
    }

    public render(): JSX.Element | null {
        return (
            <div className='home'>
                <img className='home__logo' src={`${sourcegraphContext.assetsRoot}/img/ui2/sourcegraph-head-logo.svg`} />
                <div className='home__search-box-container'>
                    <SearchBox {...this.props} />
                    <button
                        type='button'
                        className={'home__help-button' + (this.state.helpVisible ? ' home__help-button--active' : '')}
                        title={(this.state.helpVisible ? 'Hide' : 'Show') + ' help'}
                        onClick={this.toggleHelp}
                        style={{ color: this.state.helpVisible ? 'white' : 'inherit' }}
                    >
                        <HelpIcon />
                    </button>
                    <div className={'home__instructions' + (this.state.helpVisible ? ' home__instructions--open' : '')}>
                        <button type='button' className='home__instructions-close-button' onClick={this.hideHelp}><CloseIcon /></button>
                        <div className='home__instruction-row'>
                            <div className='home__explanation'>
                                Go to the repository github.com/gorilla/mux
                            </div>
                            <div className='home__instruction'>
                                gorilla/mux <kbd>tab</kbd> <kbd>enter</kbd>
                            </div>
                        </div>
                        <div className='home__instruction-row'>
                            <div className='home__explanation'>
                                Go to the file route.go in that repository
                            </div>
                            <div className='home__instruction'>
                                gorilla/mux <kbd>tab</kbd> route.go <kbd>tab</kbd> <kbd>enter</kbd>
                            </div>
                        </div>
                        <div className='home__instruction-row'>
                            <div className='home__explanation'>
                                Search a single repository for the term "route"
                            </div>
                            <div className='home__instruction'>
                                gorilla/mux <kbd>tab</kbd> route <kbd>enter</kbd>
                            </div>
                        </div>
                        <div className='home__instruction-row'>
                            <div className='home__explanation'>
                                Search multiple repositories for the term "route"
                            </div>
                            <div className='home__instruction'>
                                gorilla/mux <kbd>tab</kbd> kubernetes/kubernetes <kbd>tab</kbd> route <kbd>enter</kbd>
                            </div>
                        </div>
                        <div className='home__instruction-row'>
                            <div className='home__explanation'>
                                Search the "Kubernetes" repository group
                            </div>
                            <div className='home__instruction'>
                                Kubernetes <kbd>tab</kbd> route <kbd>enter</kbd>
                            </div>
                        </div>
                        <div className='home__instruction-row'>
                            <div className='home__explanation'>
                                Search all .go files for "route"
                            </div>
                            <div className='home__instruction'>
                                gorilla/mux <kbd>tab</kbd> *.go <kbd>tab</kbd> route <kbd>enter</kbd>
                            </div>
                        </div>
                        <div className='home__instruction-row'>
                            <div className='home__explanation'>
                                Search all files except the vendor/ directory for "session"
                            </div>
                            <div className='home__instruction'>
                                aws/aws-sdk-go <kbd>tab</kbd> !vendor/** <kbd>tab</kbd> session <kbd>enter</kbd>
                            </div>
                        </div>
                        <div className='home__instruction-row'>
                            <div className='home__explanation'>
                                Search for the regular expression "type .+ struct"
                            </div>
                            <div className='home__instruction'>
                                gorilla/mux <kbd>tab</kbd> type .+ struct <kbd>{shortcutMofidier}</kbd>+<kbd>R</kbd> <kbd>enter</kbd>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        )
    }

    public componentDidUpdate(prevProps: Props, prevState: State): void {
        if (prevState.helpVisible !== this.state.helpVisible) {
            localStorage.setItem('show-search-help', this.state.helpVisible + '')
        }
    }

    private toggleHelp = () => {
        this.setState(prevState => ({ helpVisible: !prevState.helpVisible }))
    }

    private hideHelp = () => {
        this.setState({ helpVisible: false })
    }
}
