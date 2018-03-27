import * as H from 'history'
import * as React from 'react'
import { Range } from 'vscode-languageserver-types'
import { AbsoluteRepoFile, BlobViewState } from '..'
import { Resizable } from '../../components/Resizable'
import { parseHash } from '../../util/url'
import { BlobPanel2 } from './panel/BlobPanel2'
import { BlobReferencesPanel } from './references/BlobReferencesPanel'

interface Props extends AbsoluteRepoFile {
    isLightTheme: boolean
    location: H.Location
    history: H.History
}

interface State {}

const useNewBlobPanel = localStorage.getItem('newBlobPanel') !== null

export class BlobPanel extends React.PureComponent<Props, State> {
    public render(): JSX.Element | null {
        const hash = parseHash<BlobViewState>(this.props.location.hash)
        if (hash.line === undefined || hash.viewState === undefined) {
            return null // no panel
        }

        const range: Range = {
            start: { line: hash.line, character: hash.character || 0 },
            end: {
                line: hash.endLine || hash.line,
                character: hash.endCharacter || hash.character || 0,
            },
        }

        return (
            <Resizable
                className="blob-panel--resizable"
                handlePosition="top"
                defaultSize={350}
                storageKey="blob-panel"
                element={
                    useNewBlobPanel ? (
                        <BlobPanel2
                            repoPath={this.props.repoPath}
                            commitID={this.props.commitID}
                            rev={this.props.rev}
                            viewState={hash.viewState}
                            filePath={this.props.filePath}
                            position={range.start}
                            location={this.props.location}
                            history={this.props.history}
                            isLightTheme={this.props.isLightTheme}
                        />
                    ) : (
                        <BlobReferencesPanel
                            repoPath={this.props.repoPath}
                            commitID={this.props.commitID}
                            rev={this.props.rev}
                            viewState={hash.viewState}
                            filePath={this.props.filePath}
                            position={range.start}
                            location={this.props.location}
                            history={this.props.history}
                            isLightTheme={this.props.isLightTheme}
                        />
                    )
                }
            />
        )
    }
}
