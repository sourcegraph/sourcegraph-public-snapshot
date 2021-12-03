import classNames from 'classnames'
import AlertCircleOutlineIcon from 'mdi-react/AlertCircleOutlineIcon'
import FileDocumentIcon from 'mdi-react/FileDocumentIcon'
import SearchIcon from 'mdi-react/SearchIcon'
import SlackIcon from 'mdi-react/SlackIcon'
import React from 'react'
import { Link } from 'react-router-dom'

import { GroupDetailFields } from '../../../../graphql-operations'

interface Props {
    group: GroupDetailFields
    className?: string
}

export const GroupOverviewTab: React.FunctionComponent<Props> = ({ group, className }) => (
    <div className={classNames('d-flex flex-column', className)}>
        <div className="row">
            <div className="col-md-8">hello</div>
            <div className="col-md-4">
                {group.description && <p className="mb-3">{group.description}</p>}
                <div>
                    <Link
                        to={`/search?q=context:g/${group.name}`}
                        className="d-inline-flex align-items-center btn btn-outline-secondary mb-3"
                    >
                        <SearchIcon className="icon-inline mr-1" /> Search...
                    </Link>
                    {group.readme && (
                        <div className="d-flex align-items-start">
                            <Link to={group.readme.url} className="d-flex align-items-center text-body mb-3 mr-2">
                                <FileDocumentIcon className="icon-inline mr-2" />
                                Handbook page
                            </Link>
                        </div>
                    )}
                    <Link to="#" className="d-flex align-items-center text-body mb-3">
                        <AlertCircleOutlineIcon className="icon-inline mr-2" />
                        Issues
                    </Link>
                    <Link to="#" className="d-flex align-items-center text-body mb-3">
                        <SlackIcon className="icon-inline mr-2" />
                        #extensibility-chat
                    </Link>
                    <hr className="my-3" />
                </div>
            </div>
        </div>
    </div>
)
