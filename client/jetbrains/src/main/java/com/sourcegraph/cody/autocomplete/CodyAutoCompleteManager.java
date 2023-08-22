package com.sourcegraph.cody.autocomplete;

import com.intellij.codeInsight.lookup.LookupManager;
import com.intellij.injected.editor.EditorWindow;
import com.intellij.openapi.application.ApplicationManager;
import com.intellij.openapi.diagnostic.Logger;
import com.intellij.openapi.editor.*;
import com.intellij.openapi.editor.ex.EditorEx;
import com.intellij.openapi.editor.impl.ImaginaryEditor;
import com.intellij.openapi.fileEditor.FileDocumentManager;
import com.intellij.openapi.project.Project;
import com.intellij.openapi.util.Disposer;
import com.intellij.openapi.util.Key;
import com.intellij.openapi.util.TextRange;
import com.intellij.psi.codeStyle.CommonCodeStyleSettings;
import com.intellij.util.concurrency.annotations.RequiresEdt;
import com.sourcegraph.cody.CodyCompatibility;
import com.sourcegraph.cody.agent.CodyAgent;
import com.sourcegraph.cody.agent.CodyAgentServer;
import com.sourcegraph.cody.agent.protocol.AutocompleteExecuteParams;
import com.sourcegraph.cody.autocomplete.render.*;
import com.sourcegraph.cody.statusbar.CodyAutocompleteStatus;
import com.sourcegraph.cody.statusbar.CodyAutocompleteStatusService;
import com.sourcegraph.cody.vscode.*;
import com.sourcegraph.cody.vscode.InlineAutoCompleteList;
import com.sourcegraph.common.EditorUtils;
import com.sourcegraph.config.ConfigUtil;
import com.sourcegraph.config.UserLevelConfig;
import com.sourcegraph.telemetry.GraphQlLogger;
import difflib.Delta;
import difflib.DiffUtils;
import difflib.Patch;
import java.util.List;
import java.util.Objects;
import java.util.Optional;
import java.util.concurrent.*;
import java.util.concurrent.atomic.AtomicReference;
import java.util.stream.Collectors;
import org.jetbrains.annotations.NotNull;
import org.jetbrains.annotations.Nullable;

/** Responsible for triggering and clearing inline code completions (the autocomplete feature). */
public class CodyAutoCompleteManager {
  private static final Logger logger = Logger.getInstance(CodyAutoCompleteManager.class);
  private static final Key<Boolean> KEY_EDITOR_SUPPORTED = Key.create("cody.editorSupported");
  private final ScheduledExecutorService scheduler = Executors.newSingleThreadScheduledExecutor();
  // TODO: figure out how to avoid the ugly nested `Future<CompletableFuture<T>>` type.
  private final AtomicReference<Optional<Future<CompletableFuture<Void>>>> currentJob =
      new AtomicReference<>(Optional.empty());
  private @Nullable AutocompleteTelemetry currentAutocompleteTelemetry = null;

  public static @NotNull CodyAutoCompleteManager getInstance() {
    return ApplicationManager.getApplication().getService(CodyAutoCompleteManager.class);
  }

  @RequiresEdt
  public void clearAutoCompleteSuggestions(@NotNull Editor editor) {
    // Log "suggested" event and clear current autocompletion
    Optional.ofNullable(editor.getProject())
        .ifPresent(
            p -> {
              if (currentAutocompleteTelemetry != null
                  && currentAutocompleteTelemetry.getStatus()
                      != AutocompletionStatus.TRIGGERED_NOT_DISPLAYED) {
                currentAutocompleteTelemetry.markCompletionHidden();
                GraphQlLogger.logAutocompleteSuggestedEvent(
                    p,
                    currentAutocompleteTelemetry.getLatencyMs(),
                    currentAutocompleteTelemetry.getDisplayDurationMs());
                currentAutocompleteTelemetry = null;
              }
            });

    // Cancel any running job
    cancelCurrentJob();

    // Clear any existing inline elements
    disposeInlays(editor);
  }

  @RequiresEdt
  public void disposeInlays(@NotNull Editor editor) {
    if (editor.isDisposed()) {
      return;
    }
    InlayModelUtils.getAllInlaysForEditor(editor).stream()
        .filter(inlay -> inlay.getRenderer() instanceof CodyAutoCompleteElementRenderer)
        .forEach(Disposer::dispose);
  }

  @RequiresEdt
  public boolean isEnabledForEditor(Editor editor) {
    return ConfigUtil.isCodyEnabled()
        && ConfigUtil.isCodyAutoCompleteEnabled()
        && editor != null
        && isProjectAvailable(editor.getProject())
        && isEditorSupported(editor);
  }

  /**
   * Triggers auto-complete suggestions for the given editor at the specified offset.
   *
   * @param editor The editor instance to provide autocomplete for.
   * @param offset The character offset in the editor to trigger auto-complete at.
   */
  public void triggerAutoComplete(
      @NotNull Editor editor, int offset, InlineCompletionTriggerKind triggerKind) {
    if (!isEnabledForEditor(editor)) {
      if (triggerKind.equals(InlineCompletionTriggerKind.INVOKE)) {
        logger.warn("triggered autocomplete with invalid editor " + editor);
      }
      return;
    }

    final Project project = editor.getProject();
    if (project == null) {
      logger.warn("triggered autocomplete with null project");
      return;
    }
    currentAutocompleteTelemetry = AutocompleteTelemetry.createAndMarkTriggered();
    GraphQlLogger.logCodyEvent(project, "completion", "started");

    TextDocument textDocument = new IntelliJTextDocument(editor, project);
    AutoCompleteDocumentContext autoCompleteDocumentContext =
        textDocument.getAutoCompleteContext(offset);
    // If the context has a valid completion trigger, cancel any running job
    // and asynchronously trigger the auto-complete
    if (triggerKind.equals(InlineCompletionTriggerKind.INVOKE)
        || autoCompleteDocumentContext.isCompletionTriggerValid()) { // TODO: skip this condition
      Callable<CompletableFuture<Void>> callable =
          () -> triggerAutoCompleteAsync(project, editor, offset, textDocument, triggerKind);
      // debouncing the autocomplete trigger
      cancelCurrentJob();
      this.currentJob.set(
          Optional.of(this.scheduler.schedule(callable, 20, TimeUnit.MILLISECONDS)));
    }
  }

  /** Asynchronously triggers auto-complete for the given editor and offset. */
  private CompletableFuture<Void> triggerAutoCompleteAsync(
      @NotNull Project project,
      @NotNull Editor editor,
      int offset,
      @NotNull TextDocument textDocument,
      InlineCompletionTriggerKind triggerKind) {
    CodyAgentServer server = CodyAgent.getServer(project);
    boolean isAgentAutocomplete = server != null;
    if (!isAgentAutocomplete) {
      logger.warn("Doing nothing, Agent is not running");
      return CompletableFuture.completedFuture(null);
    }

    Position position = textDocument.positionAt(offset);
    AutocompleteExecuteParams params =
        new AutocompleteExecuteParams()
            .setFilePath(
                Objects.requireNonNull(
                        FileDocumentManager.getInstance().getFile(editor.getDocument()))
                    .getPath())
            .setPosition(
                new com.sourcegraph.cody.agent.protocol.Position()
                    .setLine(position.line)
                    .setCharacter(position.character));

    CodyAutocompleteStatusService.notifyApplication(CodyAutocompleteStatus.AutocompleteInProgress);
    return server
        .autocompleteExecute(params)
        .thenAccept(
            result -> processAutocompleteResult(project, editor, offset, triggerKind, result))
        .exceptionally(
            error -> {
              logger.warn("failed autocomplete request " + params, error);
              return null;
            })
        .thenAccept(
            unused ->
                CodyAutocompleteStatusService.notifyApplication(CodyAutocompleteStatus.Ready));
  }

  private void processAutocompleteResult(
      @NotNull Project project,
      @NotNull Editor editor,
      int offset,
      InlineCompletionTriggerKind triggerKind,
      InlineAutoCompleteList result) {
    if (Thread.interrupted()) {
      if (triggerKind.equals(InlineCompletionTriggerKind.INVOKE)) {
        logger.warn("canceled autocomplete due to thread interruption");
      }
      return;
    }
    InlayModel inlayModel = editor.getInlayModel();
    Optional<InlineAutoCompleteItem> maybeItem = result.items.stream().findFirst();
    if (maybeItem.isEmpty()) {
      if (triggerKind.equals(InlineCompletionTriggerKind.INVOKE)) {
        logger.warn("explicit autocomplete returned empty suggestions");
        // NOTE(olafur): it would be nice to give the user a visual hint when this happens.
        // We don't do anything now because it's unclear what would be the most idiomatic
        // IntelliJ API to use.
      }
      return;
    }
    final InlineAutoCompleteItem item = maybeItem.get();
    ApplicationManager.getApplication()
        .invokeLater(
            () -> {
              this.clearAutoCompleteSuggestions(editor);

              if (currentAutocompleteTelemetry != null) {
                currentAutocompleteTelemetry.markCompletionDisplayed();
              }

              // Avoid displaying autocomplete when IntelliJ is already displaying
              // built-in completions. When built-in completions are visible, we can't
              // accept the Cody autocomplete with TAB because it accepts the built-in
              // completion.
              if (LookupManager.getInstance(project).getActiveLookup() != null) {
                if (triggerKind.equals(InlineCompletionTriggerKind.INVOKE)
                    || UserLevelConfig.isVerboseLoggingEnabled()) {
                  logger.warn("Skipping autocomplete because lookup is active: " + item);
                }
                return;
              }
              displayAgentAutocomplete(editor, offset, item, inlayModel, triggerKind);
            });
  }

  /**
   * Render inlay hints for unprocessed autocomplete results from the agent.
   *
   * <p>The reason we have a custom code path to render hints for agent autocompletions is because
   * we can use `insertText` directly and the `range` encloses the entire line.
   */
  private void displayAgentAutocomplete(
      @NotNull Editor editor,
      int offset,
      InlineAutoCompleteItem item,
      InlayModel inlayModel,
      InlineCompletionTriggerKind triggerKind) {
    TextRange range = EditorUtils.getTextRange(editor.getDocument(), item.range);
    String originalText = editor.getDocument().getText(range);
    String insertTextFirstLine = item.insertText.lines().findFirst().orElse("");
    String multilineInsertText =
        item.insertText.lines().skip(1).collect(Collectors.joining(System.lineSeparator()));

    // Run Myer's diff between the existing text in the document and the first line of the
    // `insertText` that is returned from the agent.
    // The diff algorithm returns a list of "deltas" that give us the minimal number of additions we
    // need to make to the document.
    Patch<String> patch = CodyAutoCompleteManager.diff(originalText, insertTextFirstLine);
    if (!patch.getDeltas().stream().allMatch(delta -> delta.getType() == Delta.TYPE.INSERT)) {
      if (triggerKind.equals(InlineCompletionTriggerKind.INVOKE)
          || UserLevelConfig.isVerboseLoggingEnabled()) {
        logger.warn("Skipping autocomplete with non-insert deltas: " + patch);
      }
      // Skip completions that need to delete or change characters in the existing document. We only
      // want completions to add changes to the document.
      return;
    }

    // Insert one inlay hint per delta in the first line.
    for (Delta<String> delta : patch.getDeltas()) {
      String text = String.join("", delta.getRevised().getLines());
      inlayModel.addInlineElement(
          range.getStartOffset() + delta.getOriginal().getPosition(),
          true,
          new CodyAutoCompleteSingleLineRenderer(
              text, item, editor, AutoCompleteRendererType.INLINE));
    }

    // Insert remaining lines of multiline completions as a single block element under the
    // (potentially false?) assumption that we don't need to compute diffs for them. My
    // understanding of multiline completions is that they are only supposed to be triggered in
    // situations where we insert a large block of code in an empty block.
    if (!multilineInsertText.isEmpty()) {
      inlayModel.addBlockElement(
          offset,
          true,
          false,
          Integer.MAX_VALUE,
          new CodyAutoCompleteBlockElementRenderer(multilineInsertText, item, editor));
    }
  }

  public static Patch<String> diff(String a, String b) {
    return DiffUtils.diff(characterList(a), characterList(b));
  }

  public static List<String> characterList(String value) {
    return value.chars().mapToObj(c -> String.valueOf((char) c)).collect(Collectors.toList());
  }

  // TODO: handle tabs in multiline autocomplete suggestions when we add them
  public static @NotNull InlineAutoCompleteItem normalizeIndentation(
      @NotNull InlineAutoCompleteItem item,
      @NotNull CommonCodeStyleSettings.IndentOptions indentOptions) {
    if (item.insertText.matches("^[\t ]*.+")) {
      String withoutLeadingWhitespace = item.insertText.stripLeading();
      String indentation =
          item.insertText.substring(
              0, item.insertText.length() - withoutLeadingWhitespace.length());
      String newIndentation = EditorUtils.tabsToSpaces(indentation, indentOptions);
      String newInsertText = newIndentation + withoutLeadingWhitespace;
      int rangeDiff = item.insertText.length() - newInsertText.length();
      Range newRange =
          item.range.withEnd(item.range.end.withCharacter(item.range.end.character - rangeDiff));
      return item.withInsertText(newInsertText).withRange(newRange);
    } else return item;
  }

  private boolean isProjectAvailable(Project project) {
    return project != null && !project.isDisposed();
  }

  private boolean isEditorSupported(@NotNull Editor editor) {
    if (editor.isDisposed()) {
      return false;
    }

    Boolean fromCache = KEY_EDITOR_SUPPORTED.get(editor);
    if (fromCache != null) {
      return fromCache;
    }

    boolean isSupported =
        isEditorInstanceSupported(editor)
            && CodyCompatibility.isSupportedProject(editor.getProject());
    KEY_EDITOR_SUPPORTED.set(editor, isSupported);
    return isSupported;
  }

  public static boolean isEditorInstanceSupported(@NotNull Editor editor) {
    return editor.getProject() != null
        && !editor.isViewer()
        && !editor.isOneLineMode()
        && !(editor instanceof EditorWindow)
        && !(editor instanceof ImaginaryEditor)
        && (!(editor instanceof EditorEx) || !((EditorEx) editor).isEmbeddedIntoDialogWrapper());
  }

  private void cancelCurrentJob() {
    // TODO: change this implementation when we avoid nested `Future<CompletableFuture<T>>`
    this.currentJob
        .get()
        .ifPresent(
            job -> {
              if (job.isDone()) {
                try {
                  job.get().cancel(true);
                } catch (ExecutionException
                    | InterruptedException
                    | CancellationException ignored) {
                }
              } else {
                // Cancelling the toplevel `Future<>` appears to cancel the nested
                // `CompletableFuture<>`.
                // Feel free to reimplement this entire method if it's causing problems because this
                // logic is not bulletproof.
                job.cancel(true);
              }
            });
    CodyAutocompleteStatusService.notifyApplication(CodyAutocompleteStatus.Ready);
  }
}
