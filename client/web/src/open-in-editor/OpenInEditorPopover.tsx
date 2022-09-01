import React, { useCallback } from 'react'

import { mdiClose } from '@mdi/js'
import { VisuallyHidden } from '@reach/visually-hidden'
import classNames from 'classnames'

import { Form } from '@sourcegraph/branded/src/components/Form'
import { Button, H3, Icon, Input, Link, Select, Text } from '@sourcegraph/wildcard'

import { isProjectPathValid } from './build-url'
import { EditorSettings } from './editor-settings'
import { supportedEditors } from './editors'

import styles from './OpenInEditorPopover.module.scss'

export interface OpenInEditorPopoverProps {
    editorSettings?: EditorSettings
    togglePopover: () => void
}

/**
 * A popover that displays a searchable list of revisions (grouped by type) for
 * the current repository.
 */
export const OpenInEditorPopover: React.FunctionComponent<React.PropsWithChildren<OpenInEditorPopoverProps>> = props => {
    const {editorSettings, togglePopover} = props

    const [selectedEditorId, setSelectedEditorId] = React.useState<string | undefined>(editorSettings?.editorId)
    const [defaultProjectPath, setDefaultProjectPath] = React.useState<string | undefined>(
        editorSettings?.projectPaths?.default
    )
    const areSettingsValid = selectedEditorId && isProjectPathValid(defaultProjectPath)

    const handleEditorChange = useCallback<React.ChangeEventHandler<HTMLSelectElement>>(event => {
        setSelectedEditorId(event.target.value)
    }, [])

    const onSubmit = useCallback<React.FormEventHandler<HTMLFormElement>>(event => {
        event.preventDefault()

        try {
            alert('Settings (not really) saved.');
        } catch {
            alert('Error saving settings.');
            // TODO: Handle errors
        }
    }, [])

    const onProjectPathChange = useCallback(
        (event: React.ChangeEvent<HTMLInputElement>): void => {
            event.preventDefault()
            setDefaultProjectPath(event.target.value)
        },
        [])

    return (
        <div className={styles.openInEditorPopover}>
            <Button className={styles.close} onClick={togglePopover}>
                <VisuallyHidden>Close</VisuallyHidden>
                <Icon svgPath={mdiClose} inline={false} aria-hidden={true} />
            </Button>
            <H3>Set your preferred editor</H3>
            <Text>
                Open this and other files directly in your editor. Set your project path and editor to get started.
                Update anytime in your user settings.
            </Text>

            <Form onSubmit={onSubmit} noValidate={true}>
                <Input
                    id="OpenInEditorForm-projectPath"
                    type="text"
                    label="Project path"
                    name="projectPath"
                    placeholder="/Users/username/projects"
                    required={true}
                    autoCorrect="off"
                    autoCapitalize="off"
                    spellCheck={false}
                    readOnly={false}
                    value={defaultProjectPath}
                    onChange={onProjectPathChange}
                    className={classNames('mr-sm-2')}
                />
                <Select
                    id="OpenInEditorForm-editor"
                    label="Editor"
                    message={
                        <>
                            Use a different editor?{' '}
                            <Link to="https://docs.sourcegraph.com/integration/open_in_editor">
                                Set up another editor
                            </Link>
                        </>
                    }
                    value={selectedEditorId}
                    onChange={handleEditorChange}
                >
                    <option value="" />
                    {[...supportedEditors]
                        .sort((a, b) => a.name.localeCompare(b.name))
                        .filter(editor => editor.id !== 'custom')
                        .map(editor => (
                            <option key={editor.id} value={editor.id}>
                                {editor.name}
                            </option>
                        ))}
                </Select>
                <Button variant="primary" type="submit" disabled={!areSettingsValid}>
                    Save
                </Button>
            </Form>
        </div>
    )
}
