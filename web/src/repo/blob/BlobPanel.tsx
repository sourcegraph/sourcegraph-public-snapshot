import * as H from 'history'
import * as React from 'react'
import { AbsoluteRepoFileRange } from '..'
import { Resizable } from '../../components/Resizable'
import { BlobReferencesPanel } from './references/BlobReferencesPanel'

interface Props extends AbsoluteRepoFileRange {
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
                    <BlobReferencesPanel
                        repoPath={this.props.repoPath}
                        commitID={this.props.commitID}
                        rev={this.props.rev}
                        referencesMode={this.props.modalMode}
                        filePath={this.props.filePath}
                        position={this.props.range.start}
                        location={this.props.location}
                        history={this.props.history}
                        isLightTheme={this.props.isLightTheme}
                    />
                }
            />
        )
    }
}
