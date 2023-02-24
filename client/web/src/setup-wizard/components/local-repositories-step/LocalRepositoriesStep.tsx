import { FC, HTMLAttributes, useState, useCallback, useRef } from 'react'

import { mdiInformationOutline, mdiPlus, mdiDelete, mdiGit } from '@mdi/js'
import classNames from 'classnames'

import { Button, Container, Icon, Input, Text, Tooltip } from '@sourcegraph/wildcard'

import { CustomNextButton } from '../setup-steps'

import styles from './LocalRepositoriesStep.module.scss'

//TODO: Node error & loading state
//TODO: Connect to api --> Validate repo selection
//TODO: Skip logic

interface LocalRepositoriesStepProps extends HTMLAttributes<HTMLDivElement> {}

export const LocalRepositoriesStep: FC<LocalRepositoriesStepProps> = props => {
    const { className, ...attributes } = props
    const fileInputReference = useRef<HTMLInputElement>(null)
    const [repos, setRepos] = useState<File[]>([])
    //store data as an edit state?? or use data.externalServices.nodes?

    // TODO: Implement BE solution for repo absolute file grab & connection
    const handleFileInputChange = useCallback(
        (event: React.ChangeEvent<HTMLInputElement>) => {
            const files = event.target.files
            if (files) {
                setRepos([...Object.values(files)])
            }
            if (!files) {
                return
            }
        },
        [repos]
    )

    const onLocalRepoSelect = useCallback(() => {
        fileInputReference.current?.click()
    }, [fileInputReference])

    return (
        <div {...attributes} className={classNames(className)}>
            <Text className="mb-2">Add your local repositories.</Text>

            <Container>
                <ul className={styles.list}>
                    {repos.length ? (
                        repos.map((codeHost, index) => (
                            <li
                                key={codeHost.lastModified}
                                className={classNames(
                                    'p-2 d-flex',
                                    styles.item,
                                    index + 1 !== repos.length && styles.itemBorder
                                )}
                            >
                                <Icon svgPath={mdiGit} aria-hidden={true} className="mt-1 mr-3" />
                                <div className="d-flex flex-column">
                                    <Text weight="medium" className="mb-0">
                                        {codeHost.name}
                                    </Text>
                                    <Text size="small" className="text-muted mb-0">
                                        {codeHost.name}
                                    </Text>
                                </div>

                                <Tooltip content="Delete code host connection" placement="right" debounce={0}>
                                    <Button
                                        variant="secondary"
                                        className={classNames('ml-auto px-2 py-0', styles.button)}
                                    >
                                        <Icon svgPath={mdiDelete} aria-label="Delete code host connection" />
                                    </Button>
                                </Tooltip>
                            </li>
                        ))
                    ) : (
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
                    )}

                    <li>
                        <Input
                            type="file"
                            multiple={true}
                            className="d-none"
                            ref={fileInputReference}
                            onChange={handleFileInputChange}
                        />
                        <Button
                            onClick={onLocalRepoSelect}
                            variant="secondary"
                            className={classNames('w-100 d-flex align-items-center', styles.button)}
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
                    </li>
                </ul>
            </Container>

            <CustomNextButton label="Skip" disabled={false} />
        </div>
    )
}
