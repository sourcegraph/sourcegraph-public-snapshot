import React from 'react'

import { ComponentLabelsFields } from '../../../../../graphql-operations'

export const ComponentLabelsSidebarItem: React.FunctionComponent<{
    component: ComponentLabelsFields
}> = ({ component }) =>
    component.labels.length > 0 ? (
        <dl>
            {component.labels.map(label => (
                <React.Fragment key={label.key}>
                    <dt>{label.key}</dt>
                    <dd>{label.values.join(', ')}</dd>
                </React.Fragment>
            ))}
        </dl>
    ) : null
