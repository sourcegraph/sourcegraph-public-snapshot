import { IChangeset } from '../../../../../../shared/src/graphql/schema'
import React from 'react'
import SourcePullIcon from 'mdi-react/SourcePullIcon'
import { changesetStatusColorClasses } from './colors'
import { Link } from '../../../../../../shared/src/components/Link'

interface Props {
    changesets: IChangeset[]
}

export const ChangesetList: React.FunctionComponent<Props> = ({ changesets }) => (
    <ul className="list-group">
        {changesets.map(changeset => (
            <li key={changeset.id} className="list-group-item d-flex pl-0 align-items-center">
                <div className="flex-shrink-0 flex-grow-0 m-2">
                    <SourcePullIcon className={`text-${changesetStatusColorClasses[changeset.state]}`} />
                </div>
                <div className="flex-fill overflow-hidden">
                    <h4 className="m-0">
                        <Link to={changeset.repository.url}>{changeset.repository.name}</Link> {changeset.title}
                    </h4>
                    <div className="text-truncate w-100">{changeset.body}</div>
                </div>
            </li>
        ))}
    </ul>
)
