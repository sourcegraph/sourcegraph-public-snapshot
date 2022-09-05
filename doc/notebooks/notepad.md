# Notepad
The notepad feature lets you quickly create notebooks as you navigate through Sourcegraph and create a notebook from those notes with one click. It is disabled by default and can be enabled on the Notebooks home page by click the "Enable notepad" button. Once enabled, the notepad component will be visible in the bottom right corner of your browser window. You can disable the notepad at any time by clicking the "Disable notepad" button (located in the same place).

!["Enable notepad"](https://storage.googleapis.com/sourcegraph-assets/notebooks/enable_notepad.png)

## Creating a Notebook

### Adding notes
The notepad currently supports adding three kinds of notes: searches, files, and file ranges. It intelligently detects the type of available notes, so if you're on a search page, it makes a search note available. If you're on a file view with no lines selected, you can add the entire file, and if you're on a file view with one or more lines selected, you can add the entire file or selected range. To select a range of lines you can use Shift + right click on the file's line numbers.

!["Add search to notepad"](https://storage.googleapis.com/sourcegraph-assets/notebooks/notepad_add_search.png)

### Annotating notes
When you create a note, you can optionally add text to it, which will also be added to the Notebook when created. On note creation, the annotation box is automatically opened and focused for ease of annotation. You can remove focus at any time by pressing Escape. The text annotation is automatically converted into a Markdown block in the Notebook, which means you can write your annotations using Markdown.

!["Annotating notes"](https://storage.googleapis.com/sourcegraph-assets/notebooks/notepad_annotations.png)

You can close a note after you add an annotation by clicking Meta/Cmd + Enter. To reopen it, click the paper icon button in the top right of the note.

You can also delete individual notes by clicking trash can icon in the note, or delete all notes by clicking the trash can icon in the bottom of the notepad, to the right of the "Create Notebook" button.

## Creating a Notebook
When you've added all your notes, click the "Create Notebook" button to create a Notebook from them. When a Notebook is created from the notepad, the notes are added in reverse order, so the first note you added is the first block in the Notebook, preceded by any annotations you added to that note (annotations appear as Markdown blocks).

When you click "Create Notebook" you will be taken to the newly-created Notebook and can edit it like any other Notebook. Add additional blocks and remove or edit existing blocks at will.

# Restoring your previous notepad session
If you close your Sourcegraph window and return, the notepad will display no notes, but you you can restore your previous session by clicking the "restore last session" button that is now visible. Even if you have added new notes since you've returned, you can restore you previous session without losing the new notes.

# Deleting all notes
Click the trash can icon in the bottom right of the notepad to delete all notes. You will be prompted to confirm deletion of all the notes. You cannot undo this action.