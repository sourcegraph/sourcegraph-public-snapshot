import randomColor from 'randomcolor'
import * as React from 'react'
import * as GQL from '../../../../shared/src/graphql/schema'
import { contrastingForegroundColor } from '../../util/contrastingForegroundColor'

/**
 * A brand icon for a project (if any), or else a default icon.
 */
export const ProjectAvatar: React.FunctionComponent<{
    project: Pick<GQL.IProject, 'name'>
    className?: string
}> = ({ project, className = '' }) => {
    const backgroundColor = randomColor({ seed: project.name }) as string
    return (
        <div
            className={`d-inline-flex align-items-center justify-content-center font-weight-bold rounded ${className}`}
            // tslint:disable-next-line: jsx-ban-props
            style={{
                backgroundColor,
                color: contrastingForegroundColor(backgroundColor),
                width: '39px',
                height: '39px',
            }}
        >
            {project.name[0].toUpperCase()}
        </div>
    )
}
