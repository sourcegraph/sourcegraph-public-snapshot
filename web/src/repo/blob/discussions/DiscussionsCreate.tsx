import * as H from 'history'
import * as React from 'react'
import { map, tap } from 'rxjs/operators'
import * as GQL from '../../../backend/graphqlschema'
import { parseHash } from '../../../util/url'
import { createThread } from './DiscussionsBackend'
import { DiscussionsInput } from './DiscussionsInput'
import { DiscussionsNavbar } from './DiscussionsNavbar'

interface Props {
    repoID: GQL.ID
    repoPath: string
    commitID: string
    rev: string | undefined
    filePath: string
    history: H.History
    location: H.Location
}

export class DiscussionsCreate extends React.PureComponent<Props> {
    public render(): JSX.Element | null {
        return (
            <div className="discussions-create">
                <DiscussionsNavbar {...this.props} />
                <DiscussionsInput submitLabel="Create discussion" onSubmit={this.onSubmit} {...this.props} />
            </div>
        )
    }

    private onSubmit = (title: string, contents: string) => {
        const lpr = parseHash(window.location.hash)

        return createThread({
            title,
            contents,
            targetRepo: {
                repository: this.props.repoID,
                path: this.props.filePath,
                branch: this.props.rev,
                revision: this.props.commitID,

                // TODO(slimsag:discussions): ASAP: capture proper selection info here
                selection: {
                    startLine: lpr.line || 0,
                    startCharacter: lpr.character || 0,
                    endLine: lpr.endLine || 0,
                    endCharacter: lpr.endCharacter || 0,
                    linesBefore: '',
                    lines: '',
                    linesAfter: '',
                },
            },
        }).pipe(
            tap(thread => {
                const location = this.props.location
                const hash = new URLSearchParams(location.hash.slice('#'.length))
                hash.set('tab', 'discussions')
                hash.set('threadID', thread.id)
                // TODO(slimsag:discussions): ASAP: focus the new thread's range
                this.props.history.push(location.pathname + location.search + '#' + hash.toString())
            }),
            map(thread => void 0)
        )
    }
}
