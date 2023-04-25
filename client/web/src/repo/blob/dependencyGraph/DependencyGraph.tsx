import { getTreeData } from './utils'

interface Props {
    repoID: string
    revision: string
    filePath: string
}

export const DependencyGraph: React.FC<Props> = props => {
    const data = getTreeData(props.filePath)
    return (
        <div>
            <code>{JSON.stringify(data, null, 2)}</code>
        </div>
    )
}
