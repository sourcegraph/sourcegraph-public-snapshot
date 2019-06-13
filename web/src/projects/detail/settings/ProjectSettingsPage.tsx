import React from 'react'
import { ExtensionsControllerProps } from '../../../../../shared/src/extensions/controller'
import { ProjectAreaContext } from '../ProjectArea'

interface Props extends Pick<ProjectAreaContext, 'project'>, ExtensionsControllerProps {}

/**
 * Displays and allows editing of a project's settings.
 */
export const ProjectSettingsPage: React.FunctionComponent<Props> = ({ project }) => (
    <div className="project-settings-page container mt-2">
        <h2>Settings</h2>
    </div>
)
