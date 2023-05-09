import { FC, useMemo } from 'react'

import { mdiFolderMultiplePlusOutline, mdiGit } from '@mdi/js'

import { Button, Icon, LoadingSpinner, PageHeader, Container } from '@sourcegraph/wildcard'

import { LocalRepository } from '../../../../graphql-operations';

import {
    useLocalPathsPicker,
    useLocalRepositories,
    useLocalRepositoriesPaths
} from '../../../../setup-wizard/components'

import styles from '../AppSettingsArea.module.scss'

export const LocalRepositoriesTab: FC = () => {
    const { paths, setPaths, loading: pathLoading } = useLocalRepositoriesPaths()
    const { repositories, loading } = useLocalRepositories({
        paths,
        skip: pathLoading || paths.length === 0
    })

    return (
        <div className={styles.content}>
            <PageHeader
                headingElement="h2"
                path={[{ text: 'Local repositories' }]}
                description="Add your local repositories"
                className="mb-3"
                actions={<PathsPickerActions onPathsChange={setPaths}/>}
            />

            <Container>
                { loading && <LoadingSpinner/> }
                { !loading && <LocalRepositoriesTree repositories={repositories}/> }
            </Container>
        </div>
    )
}

interface PathsPickerActionsProps {
    onPathsChange: (paths: string[]) => void
}

const PathsPickerActions: FC<PathsPickerActionsProps> = ({ onPathsChange }) => {
    const { callPathPicker } = useLocalPathsPicker()

    const handleClickCallPathPicker = async (): Promise<void> => {
        const paths = await callPathPicker()

        onPathsChange(paths)
    }

    return (
        <div className={styles.headerActions}>
            <Button variant="primary" onClick={handleClickCallPathPicker}>
                <Icon svgPath={mdiGit} aria-hidden={true} /> Add a repository
            </Button>
            <Button variant="primary" onClick={handleClickCallPathPicker}>
                <Icon svgPath={mdiFolderMultiplePlusOutline} aria-hidden={true} /> Add all repositories from a folder
            </Button>
        </div>
    )
}

interface LocalRepositoriesTreeProps {
   repositories: LocalRepository[]
}

const LocalRepositoriesTree: FC<LocalRepositoriesTreeProps> = ({ repositories }) => {

    const { folders, plainRepositories } = useMemo(() => {

    }, [repositories])

    return (
        <ul>

        </ul>
    )
}
