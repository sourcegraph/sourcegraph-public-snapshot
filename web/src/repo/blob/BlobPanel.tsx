import * as H from 'history'
import * as React from 'react'
import { Resizable } from '../../components/Resizable'
import { ReferencesWidget } from '../../references/ReferencesWidget'

interface Props {
    repoPath: string
    rev: string | undefined
    commitID: string
    filePath: string
    line: number
    character: number
    modalMode: 'local' | 'external' | undefined
    isLightTheme: boolean
    location: H.Location
    history: H.History
}

interface State {}

export class BlobPanel extends React.PureComponent<Props, State> {
    public render(): JSX.Element | null {
        return (
            <Resizable
                className="blob-panel--resizable"
                handlePosition="top"
                defaultSize={350}
                storageKey="blob-panel"
                element={
                    <ReferencesWidget
                        repoPath={this.props.repoPath}
                        commitID={this.props.commitID}
                        rev={this.props.rev}
                        referencesMode={this.props.modalMode}
                        filePath={this.props.filePath}
                        position={{ line: this.props.line, character: this.props.character }}
                        location={this.props.location}
                        history={this.props.history}
                        isLightTheme={this.props.isLightTheme}
                    />
                }
            />
        )
    }
}
