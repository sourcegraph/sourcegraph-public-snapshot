import * as H from 'history'
import { IPatch } from '../../../../../../shared/src/graphql/schema'
import Octicon, { Diff } from '@primer/octicons-react'
import React, { useState, useEffect, useCallback } from 'react'
import { Link } from '../../../../../../shared/src/components/Link'
import { ThemeProps } from '../../../../../../shared/src/theme'
import { Collapsible } from '../../../../components/Collapsible'
import { DiffStat } from '../../../../components/diff/DiffStat'
import { FileDiffNode } from '../../../../components/diff/FileDiffNode'
import { publishChangeset as _publishChangeset, queryPatchFileDiffs } from '../backend'
import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import ErrorIcon from 'mdi-react/ErrorIcon'
import { asError, isErrorLike } from '../../../../../../shared/src/util/errors'
import { FileDiffConnection } from '../../../../components/diff/FileDiffConnection'
import { FilteredConnectionQueryArgs } from '../../../../components/FilteredConnection'
import { Observer } from 'rxjs'

export interface PatchNodeProps extends ThemeProps {
    node: IPatch
    campaignUpdates?: Pick<Observer<void>, 'next'>
    history: H.History
    location: H.Location
    /** Shows the publish button */
    enablePublishing: boolean
}

export const PatchNode: React.FunctionComponent<PatchNodeProps> = ({
    node,
    campaignUpdates,
    isLightTheme,
    history,
    location,
    enablePublishing,
}) => {
    const [isPublishing, setIsPublishing] = useState<boolean | Error>(false)
    useEffect(() => {
        setIsPublishing(node.publicationEnqueued)
    }, [node.publicationEnqueued])

    const publishChangeset: React.MouseEventHandler = async () => {
        try {
            setIsPublishing(true)
            await _publishChangeset(node.id)
            if (campaignUpdates) {
                // trigger campaign update to refetch on new state
                campaignUpdates.next()
            }
        } catch (error) {
            setIsPublishing(asError(error))
        }
    }
    const fileDiffs = node.diff?.fileDiffs

    const changesetNodeRow = (
        <div className="d-flex align-items-center m-1 ml-2">
            <div className="changeset-node__content flex-fill">
                <div className="d-flex flex-column">
                    <div>
                        <Octicon icon={Diff} className="icon-inline mr-2" />
                        <strong>
                            <Link
                                to={node.repository.url}
                                className="text-muted"
                                target="_blank"
                                rel="noopener noreferrer"
                            >
                                {node.repository.name}
                            </Link>
                        </strong>
                    </div>
                </div>
            </div>
            <div className="flex-shrink-0 flex-grow-0 ml-1 align-items-end">
                {fileDiffs && <DiffStat {...fileDiffs.diffStat} expandedCounts={true} />}
            </div>
            {enablePublishing && (
                <>
                    {isErrorLike(isPublishing) && <ErrorIcon data-tooltip={isPublishing.message} className="ml-2" />}
                    <button
                        type="button"
                        className="flex-shrink-0 flex-grow-0 btn btn-sm btn-secondary ml-2"
                        disabled={!isErrorLike(isPublishing) && !!isPublishing}
                        onClick={publishChangeset}
                    >
                        {!isErrorLike(isPublishing) && !!isPublishing && <LoadingSpinner className="icon-inline" />}{' '}
                        {!isErrorLike(isPublishing) && !!isPublishing ? 'Publishing' : 'Publish'}
                    </button>
                </>
            )}
        </div>
    )

    /** Fetches the file diffs for the changeset */
    const queryFileDiffs = useCallback(
        (args: FilteredConnectionQueryArgs) => queryPatchFileDiffs(node.id, { ...args, isLightTheme }),
        [node.id, isLightTheme]
    )

    return (
        <li className="list-group-item e2e-changeset-node">
            {fileDiffs ? (
                <Collapsible
                    titleClassName="changeset-node__content flex-fill"
                    expandedButtonClassName="mb-3"
                    title={changesetNodeRow}
                    wholeTitleClickable={false}
                >
                    <FileDiffConnection
                        listClassName="list-group list-group-flush"
                        noun="changed file"
                        pluralNoun="changed files"
                        queryConnection={queryFileDiffs}
                        nodeComponent={FileDiffNode}
                        nodeComponentProps={{
                            history,
                            location,
                            isLightTheme,
                            persistLines: false,
                            lineNumbers: true,
                        }}
                        updateOnChange={node.repository.id}
                        defaultFirst={15}
                        hideSearch={true}
                        noSummaryIfAllNodesVisible={true}
                        history={history}
                        location={location}
                        useURLQuery={false}
                        cursorPaging={true}
                    />
                </Collapsible>
            ) : (
                <div className="changeset-node__content changeset-node__content--no-collapse flex-fill">
                    {changesetNodeRow}
                </div>
            )}
        </li>
    )
}
