import * as H from 'history'
import * as React from 'react'
import { Range } from 'vscode-languageserver-types'
import { AbsoluteRepoFile } from '..'
import { Resizable } from '../../components/Resizable'
import { parseHash } from '../../util/url'
import { BlobReferencesPanel } from './references/BlobReferencesPanel'

interface Props extends AbsoluteRepoFile {
    isLightTheme: boolean
    location: H.Location
    history: H.History
}

interface State {}

export class BlobPanel extends React.PureComponent<Props, State> {
    public render(): JSX.Element | null {
        const hash = parseHash(this.props.location.hash)
        if (hash.line === undefined || hash.modal === undefined) {
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
                    <BlobReferencesPanel
                        repoPath={this.props.repoPath}
                        commitID={this.props.commitID}
                        rev={this.props.rev}
                        referencesMode={hash.modalMode}
                        filePath={this.props.filePath}
                        position={range.start}
                        location={this.props.location}
                        history={this.props.history}
                        isLightTheme={this.props.isLightTheme}
                    />
                }
            />
        )
    }
}
