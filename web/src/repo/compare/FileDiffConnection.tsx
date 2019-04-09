import * as React from 'react'
import { ViewComponentData } from '../../../../shared/src/api/client/model'
import { ExtensionsControllerProps } from '../../../../shared/src/extensions/controller'
import * as GQL from '../../../../shared/src/graphql/schema'
import { getModeFromPath } from '../../../../shared/src/languages'
import { ErrorLike, isErrorLike } from '../../../../shared/src/util/errors'
import { Connection, FilteredConnection } from '../../components/FilteredConnection'
import { FileDiffNodeProps } from './FileDiffNode'

class FilteredFileDiffConnection extends FilteredConnection<
    GQL.IFileDiff,
    Pick<
        FileDiffNodeProps,
        | 'base'
        | 'head'
        | 'lineNumbers'
        | 'className'
        | 'platformContext'
        | 'location'
        | 'history'
        | 'hoverifier'
        | 'extensionsController'
    >
> {}

type Props = FilteredFileDiffConnection['props'] & ExtensionsControllerProps

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

        const visibleViewComponents: ViewComponentData[] = []
        if (fileDiffsOrError && !isErrorLike(fileDiffsOrError)) {
            for (const fileDiff of fileDiffsOrError.nodes) {
                if (fileDiff.oldPath) {
                    visibleViewComponents.push({
                        type: 'CodeEditor',
                        item: {
                            uri: `git://${nodeProps.base.repoName}?${nodeProps.base.commitID}#${fileDiff.oldPath}`,
                            languageId: getModeFromPath(fileDiff.oldPath),
                            text: dummyText,
                        },
                        selections: [],
                        isActive: false, // HACK: arbitrarily say that the base is inactive. TODO: support diffs first-class
                    })
                }
                if (fileDiff.newPath) {
                    visibleViewComponents.push({
                        type: 'CodeEditor',
                        item: {
                            uri: `git://${nodeProps.head.repoName}?${nodeProps.head.commitID}#${fileDiff.newPath}`,
                            languageId: getModeFromPath(fileDiff.newPath),
                            text: dummyText,
                        },
                        selections: [],
                        isActive: true,
                    })
                }
            }
        }
        this.props.extensionsController.services.editor.model.next({ visibleViewComponents })
    }
}
