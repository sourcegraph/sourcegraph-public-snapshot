import { lazyComponent } from '@sourcegraph/shared/src/util/lazyComponent'

interface Props {
    repoID: string
    revision: string
    filePath: string
}

const DependencyGraph = lazyComponent(() => import('./DependencyGraph'), 'DependencyGraph')

export const DependencyGraphPanel: React.FC<Props> = props => {
    return (
        <div>
            <DependencyGraph {...props} />
        </div>
    )
}
