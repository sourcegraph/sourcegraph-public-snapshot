import React from 'react'

import { InstanceData } from '../model'

export const Instance: React.FunctionComponent<{
    instance: InstanceData
    tag?: 'li'
}> = ({ instance, tag: Tag = 'li' }) => (
    <Tag className="instance">
        <div className="container">
            <h2>
                <a href={instance.url}>{instance.title}</a> <code>{instance.id}</code>
            </h2>
        </div>
    </Tag>
)
