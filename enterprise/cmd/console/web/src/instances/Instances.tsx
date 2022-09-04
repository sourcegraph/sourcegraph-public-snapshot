import './Instances.css'

import React from 'react'

import { InstanceData } from '../model'
import { Instance } from './Instance'

export const Instances: React.FunctionComponent<{
    instances: InstanceData[] | undefined
    tag?: 'main'
    className?: string
}> = ({ instances, tag: Tag = 'main', className }) => (
    <Tag className={className}>
        {instances === undefined ? (
            <div className="container">
                <p>Loading...</p>
            </div>
        ) : instances.length === 0 ? (
            <div className="container">
                <p>No instances found.</p>
            </div>
        ) : (
            <ol className="instance-groups">
                {instances.map((instance, i) => (
                    <Instance key={i} instance={instance} tag="li" />
                ))}
            </ol>
        )}
    </Tag>
)
