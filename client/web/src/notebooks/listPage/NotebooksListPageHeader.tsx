import React, { useCallback, useRef, useState } from 'react'

import { mdiChevronDown } from '@mdi/js'
import { VisuallyHidden } from '@reach/visually-hidden'
import * as uuid from 'uuid'

import type { ErrorLike } from '@sourcegraph/common'
import { useTemporarySetting } from '@sourcegraph/shared/src/settings/temporary/useTemporarySetting'
import { TelemetryV2Props } from '@sourcegraph/shared/src/telemetry'
import type { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import {
    Link,
    Button,
    Menu,
    ButtonGroup,
    MenuButton,
    Position,
    MenuList,
    MenuItem,
    Input,
    Modal,
    Icon,
} from '@sourcegraph/wildcard'

import type { AuthenticatedUser } from '../../auth'
import type { CreateNotebookVariables } from '../../graphql-operations'
import { EnterprisePageRoutes } from '../../routes.constants'
import { NotepadIcon } from '../../search/Notepad'
import { blockToGQLInput } from '../serialize'
import { convertMarkdownToBlocks } from '../serialize/convertMarkdownToBlocks'

import { NOTEPAD_CTA_ID, NotepadCTA } from './NotepadCta'

import styles from './NotebooksListPageHeader.module.scss'

const LOADING = 'loading' as const

const INVALID_IMPORT_FILE_ERROR = new Error(
    'Cannot read the imported file. Check that the imported file is a Markdown-formatted text file.'
)

const MAX_FILE_SIZE_IN_BYTES = 1000 * 1000 // 1MB

interface NotebooksListPageHeaderProps extends TelemetryProps, TelemetryV2Props {
    authenticatedUser: AuthenticatedUser
    importNotebook: (notebook: CreateNotebookVariables['notebook']) => void
    setImportState: (state: typeof LOADING | ErrorLike | undefined) => void
}

export const NotebooksListPageHeader: React.FunctionComponent<
    React.PropsWithChildren<NotebooksListPageHeaderProps>
> = ({ authenticatedUser, telemetryService, telemetryRecorder, setImportState, importNotebook }) => {
    const fileInputReference = useRef<HTMLInputElement>(null)

    const onImportMenuItemSelect = useCallback(() => {
        telemetryService.log('SearchNotebookImportMarkdownNotebookButtonClick')
        telemetryRecorder.recordEvent('SearchNotebookImportMarkdownNotebookButton', 'clicked')
        // Open the system file picker.
        fileInputReference.current?.click()
    }, [fileInputReference, telemetryService, telemetryRecorder])

    const onFileLoad = useCallback(
        (event: ProgressEvent<FileReader>, fileName: string): void => {
            if (typeof event.target?.result !== 'string') {
                setImportState(INVALID_IMPORT_FILE_ERROR)
                return
            }
            const blocks = convertMarkdownToBlocks(event.target.result).map(block =>
                blockToGQLInput({ id: uuid.v4(), ...block })
            )
            const title = fileName.split('.snb.md')[0].trim() || 'New Notebook'
            importNotebook({ title, blocks, public: false, namespace: authenticatedUser.id })
        },
        [authenticatedUser, importNotebook, setImportState]
    )

    const onFileInputChange = useCallback(
        (event: React.ChangeEvent<HTMLInputElement>) => {
            const files = event.target.files
            if (!files || files.length !== 1) {
                setImportState(INVALID_IMPORT_FILE_ERROR)
                return
            }

            if (files[0].size > MAX_FILE_SIZE_IN_BYTES) {
                setImportState(new Error('File too large. Maximum allowed file size is 1MB.'))
                return
            }

            setImportState(LOADING)

            const reader = new FileReader()
            reader.addEventListener('load', event => onFileLoad(event, files[0].name))
            reader.readAsText(files[0])
        },
        [setImportState, onFileLoad]
    )

    return (
        <>
            <ToggleNotepadButton
                telemetryService={telemetryService}
                telemetryRecorder={telemetryRecorder}
                className="mr-2 d-none d-md-inline"
            />
            {/* The file upload input has to always be present in the DOM, otherwise the upload process
            does not complete when the menu below closes.  */}
            <Input
                type="file"
                className="d-none"
                ref={fileInputReference}
                accept=".md"
                onChange={onFileInputChange}
                data-testid="import-markdown-notebook-file-input"
            />
            <Menu>
                <ButtonGroup>
                    <Button to={EnterprisePageRoutes.NotebookCreate} variant="primary" as={Link}>
                        Create notebook
                    </Button>
                    <MenuButton variant="primary" className={styles.dropdownButton}>
                        <Icon aria-hidden={true} svgPath={mdiChevronDown} />
                        <VisuallyHidden>Actions</VisuallyHidden>
                    </MenuButton>
                </ButtonGroup>
                <MenuList position={Position.bottomEnd}>
                    <MenuItem className={styles.menuItem} onSelect={onImportMenuItemSelect}>
                        Import Markdown notebook
                    </MenuItem>
                </MenuList>
            </Menu>
        </>
    )
}

export const NOTEPAD_ENABLED_EVENT = 'SearchNotepadEnabled'
const NOTEPAD_DISABLED_EVENT = 'SearchNotepadDisabled'

const ToggleNotepadButton: React.FunctionComponent<
    React.PropsWithChildren<TelemetryProps & TelemetryV2Props & { className?: string }>
> = ({ telemetryService, telemetryRecorder, className }) => {
    const [notepadEnabled, setNotepadEnabled] = useTemporarySetting('search.notepad.enabled')
    const [ctaSeen, setCTASeen] = useTemporarySetting('search.notepad.ctaSeen')
    const [showCTA, setShowCTA] = useState(false)

    function onClick(): void {
        if (!notepadEnabled && !ctaSeen) {
            setShowCTA(true)
        } else {
            setNotepadEnabled(enabled => {
                // `enabled` is the old state so we have to log the "opposite"
                // event
                telemetryService.log(enabled ? NOTEPAD_DISABLED_EVENT : NOTEPAD_ENABLED_EVENT)
                telemetryRecorder.recordEvent(enabled ? NOTEPAD_DISABLED_EVENT : NOTEPAD_ENABLED_EVENT, 'clicked')
                return !enabled
            })
        }
    }

    function onEnableFromCTA(): void {
        telemetryService.log(NOTEPAD_ENABLED_EVENT)
        setNotepadEnabled(true)
        setShowCTA(false)
        setCTASeen(true)
    }

    function onCancelFromCTA(): void {
        // We only mark the CTA as "seen" when the user enables the notepad from it
        setShowCTA(false)
    }

    return (
        <>
            <Button variant="secondary" type="button" onClick={onClick} className={className}>
                <NotepadIcon /> {notepadEnabled ? 'Disable' : 'Enable'} notepad
            </Button>
            {showCTA && (
                <Modal aria-labelledby={NOTEPAD_CTA_ID}>
                    <NotepadCTA onEnable={onEnableFromCTA} onCancel={onCancelFromCTA} />
                </Modal>
            )}
        </>
    )
}
