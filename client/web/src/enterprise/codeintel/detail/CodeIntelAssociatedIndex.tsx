import ChevronRightIcon from 'mdi-react/ChevronRightIcon'
import React, { FunctionComponent } from 'react'

import { Link } from '@sourcegraph/shared/src/components/Link'

import { LsifUploadFields } from '../../../graphql-operations'
import { CodeIntelState } from '../shared/CodeIntelState'
import { CodeIntelUploadOrIndexLastActivity } from '../shared/CodeIntelUploadOrIndexLastActivity'

export interface CodeIntelAssociatedIndexProps {
    node: LsifUploadFields
    now?: () => Date
}

export const CodeIntelAssociatedIndex: FunctionComponent<CodeIntelAssociatedIndexProps> = ({ node, now }) =>
    node.associatedIndex && node.projectRoot ? (
        <>
            <div className="list-group position-relative">
                <div className="codeintel-associated-index__grid mb-3">
                    <span className="codeintel-associated-index__separator" />

                    <div className="d-flex flex-column codeintel-associated-index__information">
                        <div className="m-0">
                            <h3 className="m-0 d-block d-md-inline">This upload was created by an auto-indexing job</h3>
                        </div>

                        <div>
                            <small className="text-mute">
                                <CodeIntelUploadOrIndexLastActivity
                                    node={{ ...node.associatedIndex, uploadedAt: null }}
                                    now={now}
                                />
                            </small>
                        </div>
                    </div>

                    <span className="d-none d-md-inline codeintel-associated-index__state">
                        <CodeIntelState node={node.associatedIndex} className="d-flex flex-column align-items-center" />
                    </span>
                    <span>
                        <Link
                            to={`/${node.projectRoot.repository.name}/-/settings/code-intelligence/indexes/${node.associatedIndex.id}`}
                        >
                            <ChevronRightIcon />
                        </Link>
                    </span>

                    <span className="codeintel-associated-index__separator" />
                </div>
            </div>
        </>
    ) : (
        <></>
    )
