import './InstanceList.css'

import React from 'react'

import { InstanceData } from '../model'
import { Instance } from './Instance'

export const InstanceList: React.FunctionComponent<{
    instances: InstanceData[]
    tag?: 'main'
    className?: string
}> = ({ instances, tag: Tag = 'main', className }) => (
    <Tag className={className}>
        {instances.length === 0 ? (
            <div className="container">
                <p>No instances found.</p>
            </div>
        ) : (
            <ol className="instances">
                {instances.map((instance, i) => (
                    <Instance key={i} instance={instance} tag="li" />
                ))}
            </ol>
        )}
    </Tag>
)
