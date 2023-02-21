import { FC, HTMLAttributes } from 'react'

import { mdiInformationOutline, mdiPlus } from '@mdi/js'
import classNames from 'classnames'

import { Button, Container, Icon, Input, Text } from '@sourcegraph/wildcard'

import { CustomNextButton } from '../setup-steps'

import styles from './LocalRepositoriesStep.module.scss'

//TODO: Skip logic
//TODO: Error state
//TODO: Connect to api

interface LocalRepositoriesStepProps extends HTMLAttributes<HTMLDivElement> {}

export const LocalRepositoriesStep: FC<LocalRepositoriesStepProps> = props => {
    const { className, ...attributes } = props

    const handleFileInputChange = () => {
        //TODO: Customize file input (ref <NotebooksListPageHeader /> --> onImportMenuItemSelect)
        // Selected file adds new node (loading / api call)
        // New file input available
        //TODO: Validate repo selected
    }

    return (
        <div {...attributes} className={classNames(className)}>
            <Text className="mb-2">Add your local repositories.</Text>

            <Container>
                <Text weight="bold" className="d-flex align-items-center font-weight-bold text-muted">
                    <Icon
                        svgPath={mdiInformationOutline}
                        className="mr-2 mx-2"
                        inline={false}
                        aria-hidden={true}
                        height={22}
                        width={22}
                    />
                    To get started, add at least one local repository to Sourcegraph.
                </Text>

                <Button
                    variant="secondary"
                    className={classNames('w-100 d-flex align-items-center', styles.addRepositoryButton)}
                    outline={true}
                >
                    <Icon svgPath={mdiPlus} aria-hidden={true} height={26} width={26} />
                    <div className="ml-2">
                        <Text weight="medium" className="text-left mb-0">
                            Add existing local repositories.
                        </Text>
                        <Text size="small" className="text-muted text-left mb-0">
                            Multiple folders can be selected at once.
                        </Text>
                    </div>
                </Button>

                <Input type="file" onChange={handleFileInputChange} />
            </Container>

            <CustomNextButton label="Skip" disabled={false} />
        </div>
    )
}
