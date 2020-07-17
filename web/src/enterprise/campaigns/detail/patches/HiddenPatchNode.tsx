import React from 'react'
import Octicon, { Diff } from '@primer/octicons-react'

export const HiddenPatchNode: React.FunctionComponent<{}> = () => (
    <li className="list-group-item test-changeset-node">
        <div className="changeset-node__content changeset-node__content--no-collapse flex-fill">
            <div className="d-flex align-items-center m-1 ml-2">
                <div className="changeset-node__content flex-fill">
                    <div className="d-flex flex-column">
                        <div>
                            <Octicon icon={Diff} className="icon-inline text-muted mr-2" />
                            <strong className="text-muted">Patch in a private repository</strong>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    </li>
)
