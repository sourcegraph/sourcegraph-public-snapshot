import * as React from 'react'
import { TextDocumentItem } from 'sourcegraph/module/client/types/textDocument'
import * as GQL from '../../backend/graphqlschema'
import { Connection, FilteredConnection } from '../../components/FilteredConnection'
import { ExtensionsDocumentsProps } from '../../extensions/environment/ExtensionsEnvironment'
import { getModeFromPath } from '../../util'
import { ErrorLike, isErrorLike } from '../../util/errors'
import { FileDiffNodeProps } from './FileDiffNode'

class FilteredFileDiffConnection extends FilteredConnection<
    GQL.IFileDiff,
    Pick<
        FileDiffNodeProps,
        | 'base'
        | 'head'
        | 'lineNumbers'
        | 'className'
        | 'extensions'
        | 'location'
        | 'history'
        | 'hoverifier'
        | 'extensionsController'
    >
> {}

type Props = FilteredFileDiffConnection['props'] & ExtensionsDocumentsProps

/**
 * Displays a list of file diffs.
 */
export class FileDiffConnection extends React.PureComponent<Props> {
    public render(): JSX.Element | null {
        return <FilteredFileDiffConnection {...this.props} onUpdate={this.onUpdate} />
    }

    private onUpdate = (fileDiffsOrError: Connection<GQL.IFileDiff> | ErrorLike | undefined) => {
        const nodeProps = this.props.nodeComponentProps!

        // TODO(sqs): This reports to extensions that these files are empty. This is wrong, but we don't have any
        // easy way to get the files' full contents here (and doing so would be very slow). Improve the extension
        // API's support for diffs.
        const dummyText = ''

        const visibleTextDocuments: TextDocumentItem[] = []
        if (fileDiffsOrError && !isErrorLike(fileDiffsOrError)) {
            for (const fileDiff of fileDiffsOrError.nodes) {
                if (fileDiff.oldPath) {
                    visibleTextDocuments.push({
                        uri: `git://${nodeProps.base.repoPath}?${nodeProps.base.commitID}#${fileDiff.oldPath}`,
                        languageId: getModeFromPath(fileDiff.oldPath),
                        text: dummyText,
                    })
                }
                if (fileDiff.newPath) {
                    visibleTextDocuments.push({
                        uri: `git://${nodeProps.head.repoPath}?${nodeProps.head.commitID}#${fileDiff.newPath}`,
                        languageId: getModeFromPath(fileDiff.newPath),
                        text: dummyText,
                    })
                }
            }
        }
        this.props.extensionsOnVisibleTextDocumentsChange(visibleTextDocuments)
    }
}
