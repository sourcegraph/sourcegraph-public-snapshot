import { upperFirst } from 'lodash'
import React from 'react'
import { Link } from 'react-router-dom'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { threadNoun } from '../../util'
import { ThreadAreaContext } from '../ThreadArea'

interface Props extends Pick<ThreadAreaContext, 'project'> {
    thread: GQL.IDiscussionThread

    /** The project containing the thread. */
    project: Pick<GQL.IProject, 'id' | 'name' | 'url'> | null

    areaURL: string
    className?: string
}

/**
 * The breadcrumbs for a thread.
 */
export const ThreadBreadcrumbs: React.FunctionComponent<Props> = ({ thread, project, areaURL, className = '' }) => (
    <nav className={`d-flex align-items-center ${className}`} aria-label="breadcrumb">
        <ol className="breadcrumb">
            {project && (
                <li className="breadcrumb-item">
                    <Link to={project.url}>{project.name}</Link>
                </li>
            )}
            <li className="breadcrumb-item">
                <Link
                    to={`${project ? project.url : ''}${thread.type === GQL.ThreadType.CHECK ? '/checks' : '/threads'}`}
                >
                    {upperFirst(threadNoun(thread.type, true))}
                </Link>
            </li>
            <li className="breadcrumb-item active font-weight-bold">
                <Link to={areaURL}>#{thread.idWithoutKind}</Link>
            </li>
        </ol>
    </nav>
)
