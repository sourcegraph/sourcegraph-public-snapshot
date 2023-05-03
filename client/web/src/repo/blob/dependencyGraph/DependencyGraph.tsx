import ReactFlow, { Controls, Background } from 'reactflow'

import 'reactflow/dist/style.css'

import { getTreeData } from './utils'

import styles from './DependencyGraph.module.scss'

interface Props {
    repoID: string
    revision: string
    filePath: string
}

export const DependencyGraph: React.FC<Props> = props => {
    const data = getTreeData(props.filePath)

    const nodes = [
        {
            id: '1',
            position: { x: 0, y: 0 },
            data: { label: 'Hello' },
            type: 'input',
        },
        {
            id: '2',
            position: { x: 100, y: 100 },
            data: { label: 'World' },
        },
    ]

    return (
        <div className={styles.container}>
            <code>{JSON.stringify(data, null, 2)}</code>
            <div style={{ height: '100%' }}>
                <ReactFlow nodes={nodes}>
                    <Background />
                    <Controls />
                </ReactFlow>
            </div>
        </div>
    )
}
