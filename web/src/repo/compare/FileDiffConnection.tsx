import * as React from 'react'
import { Omit } from 'utility-types'
import { ExtensionsControllerProps } from '../../../../shared/src/extensions/controller'
import * as GQL from '../../../../shared/src/graphql/schema'
import { getModeFromPath } from '../../../../shared/src/languages'
import { ErrorLike, isErrorLike } from '../../../../shared/src/util/errors'
import { Connection, FilteredConnection } from '../../components/FilteredConnection'
import { FileDiffNodeProps } from './FileDiffNode'

class FilteredFileDiffConnection extends FilteredConnection<GQL.IFileDiff, Omit<FileDiffNodeProps, 'node'>> {}

type Props = FilteredFileDiffConnection['props'] & ExtensionsControllerProps

/**
 * Displays a list of file diffs.
 */
export class FileDiffConnection extends React.PureComponent<Props> {
    public render(): JSX.Element | null {
        return <FilteredFileDiffConnection {...this.props} onUpdate={this.onUpdate} />
    }

    private onUpdate = (fileDiffsOrError: Connection<GQL.IFileDiff> | ErrorLike | undefined): void => {
        const nodeProps = this.props.nodeComponentProps!

        // TODO(sqs): This reports to extensions that these files are empty. This is wrong, but we don't have any
        // easy way to get the files' full contents here (and doing so would be very slow). Improve the extension
        // API's support for diffs.
        const dummyText = ''

        this.props.extensionsController.services.editor.removeAllEditors()

        if (fileDiffsOrError && !isErrorLike(fileDiffsOrError)) {
            for (const fileDiff of fileDiffsOrError.nodes) {
                if (fileDiff.oldPath) {
                    const uri = `git://${nodeProps.base.repoName}?${nodeProps.base.commitID}#${fileDiff.oldPath}`
                    if (!this.props.extensionsController.services.model.hasModel(uri)) {
                        this.props.extensionsController.services.model.addModel({
                            uri,
                            languageId: getModeFromPath(fileDiff.oldPath),
                            text: dummyText,
                        })
                    }
                    this.props.extensionsController.services.editor.addEditor({
                        type: 'CodeEditor',
                        resource: uri,
                        selections: [],
                        isActive: false, // HACK: arbitrarily say that the base is inactive. TODO: support diffs first-class
                    })
                }
                if (fileDiff.newPath) {
                    const uri = `git://${nodeProps.head.repoName}?${nodeProps.head.commitID}#${fileDiff.newPath}`
                    if (!this.props.extensionsController.services.model.hasModel(uri)) {
                        this.props.extensionsController.services.model.addModel({
                            uri,
                            languageId: getModeFromPath(fileDiff.newPath),
                            text: dummyText,
                        })
                    }
                    this.props.extensionsController.services.editor.addEditor({
                        type: 'CodeEditor',
                        resource: uri,
                        selections: [],
                        isActive: true,
                    })
                }
            }
        }
    }
}
