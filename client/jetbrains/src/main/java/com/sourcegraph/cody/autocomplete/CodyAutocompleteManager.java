package com.sourcegraph.cody.autocomplete;

import com.intellij.codeInsight.lookup.LookupManager;
import com.intellij.openapi.application.ApplicationManager;
import com.intellij.openapi.diagnostic.Logger;
import com.intellij.openapi.editor.*;
import com.intellij.openapi.fileEditor.FileDocumentManager;
import com.intellij.openapi.project.Project;
import com.intellij.openapi.util.Disposer;
import com.intellij.openapi.util.TextRange;
import com.intellij.openapi.vfs.VirtualFile;
import com.intellij.psi.codeStyle.CommonCodeStyleSettings;
import com.intellij.util.concurrency.annotations.RequiresEdt;
import com.sourcegraph.cody.agent.CodyAgent;
import com.sourcegraph.cody.agent.CodyAgentServer;
import com.sourcegraph.cody.agent.protocol.AutocompleteExecuteParams;
import com.sourcegraph.cody.autocomplete.render.*;
import com.sourcegraph.cody.statusbar.CodyAutocompleteStatus;
import com.sourcegraph.cody.statusbar.CodyAutocompleteStatusService;
import com.sourcegraph.cody.vscode.*;
import com.sourcegraph.cody.vscode.InlineAutocompleteList;
import com.sourcegraph.config.ConfigUtil;
import com.sourcegraph.config.UserLevelConfig;
import com.sourcegraph.telemetry.GraphQlLogger;
import com.sourcegraph.utils.CodyEditorUtil;
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
public class CodyAutocompleteManager {
  private static final Logger logger = Logger.getInstance(CodyAutocompleteManager.class);
  private final ScheduledExecutorService scheduler = Executors.newSingleThreadScheduledExecutor();
  private final AtomicReference<CancellationToken> currentJob =
      new AtomicReference<>(new CancellationToken());

  public @Nullable AutocompleteTelemetry getCurrentAutocompleteTelemetry() {
    return currentAutocompleteTelemetry;
  }

  private @Nullable AutocompleteTelemetry currentAutocompleteTelemetry = null;

  public static @NotNull CodyAutocompleteManager getInstance() {
    return ApplicationManager.getApplication().getService(CodyAutocompleteManager.class);
  }

  /**
   * Clears any already rendered autocomplete suggestions for the given editor and cancels any
   * pending ones.
   *
   * @param editor the editor to clear autocomplete suggestions for
   */
  @RequiresEdt
  public void clearAutocompleteSuggestions(@NotNull Editor editor) {
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
                    currentAutocompleteTelemetry.getDisplayDurationMs(),
                    currentAutocompleteTelemetry.params());
                currentAutocompleteTelemetry = null;
              }
            });

    // Cancel any running job
    cancelCurrentJob();

    // Clear any existing inline elements
    disposeInlays(editor);
  }

  /**
   * Clears any already rendered autocomplete suggestions for all open editors and cancels any
   * pending ones.
   */
  @RequiresEdt
  public void clearAutocompleteSuggestionsForAllProjects() {
    CodyEditorUtil.getAllOpenEditors().forEach(this::clearAutocompleteSuggestions);
  }

  @RequiresEdt
  public void clearAutocompleteSuggestionsForLanguageIds(List<String> languageIds) {
    CodyEditorUtil.getAllOpenEditors().stream()
        .filter(
            e ->
                Optional.ofNullable(CodyEditorUtil.getLanguage(e))
                    .map(l -> languageIds.contains(l.getID()))
                    .orElse(false))
        .forEach(this::clearAutocompleteSuggestions);
  }

  @RequiresEdt
  public void disposeInlays(@NotNull Editor editor) {
    if (editor.isDisposed()) {
      return;
    }
    InlayModelUtils.getAllInlaysForEditor(editor).stream()
        .filter(inlay -> inlay.getRenderer() instanceof CodyAutocompleteElementRenderer)
        .forEach(Disposer::dispose);
  }

  /**
   * Triggers auto-complete suggestions for the given editor at the specified offset.
   *
   * @param editor The editor instance to provide autocomplete for.
   * @param offset The character offset in the editor to trigger auto-complete at.
   */
  public void triggerAutocomplete(
      @NotNull Editor editor, int offset, InlineCompletionTriggerKind triggerKind) {
    boolean isTriggeredManually = triggerKind.equals(InlineCompletionTriggerKind.INVOKE);
    if (!ConfigUtil.isCodyEnabled()) return;
    else if (!CodyEditorUtil.isEditorValidForAutocomplete(editor)) {
      if (isTriggeredManually) logger.warn("triggered autocomplete with invalid editor " + editor);
      return;
    } else if (!isTriggeredManually
        && !CodyEditorUtil.isImplicitAutocompleteEnabledForEditor(editor)) return;

    final Project project = editor.getProject();
    if (project == null) {
      logger.warn("triggered autocomplete with null project");
      return;
    }
    currentAutocompleteTelemetry = AutocompleteTelemetry.createAndMarkTriggered();

    TextDocument textDocument = new IntelliJTextDocument(editor, project);
    AutocompleteDocumentContext autoCompleteDocumentContext =
        textDocument.getAutocompleteContext(offset);
    // If the context has a valid completion trigger, cancel any running job
    // and asynchronously trigger the auto-complete
    if (triggerKind.equals(InlineCompletionTriggerKind.INVOKE)
        || autoCompleteDocumentContext.isCompletionTriggerValid()) { // TODO: skip this condition
      CancellationToken cancellationToken = new CancellationToken();
      Callable<CompletableFuture<Void>> callable =
          () ->
              triggerAutocompleteAsync(
                  project, editor, offset, textDocument, triggerKind, cancellationToken);
      ScheduledFuture<CompletableFuture<Void>> scheduledAutocomplete =
          this.scheduler.schedule(callable, 20, TimeUnit.MILLISECONDS);
      cancellationToken.onCancellationRequested(() -> scheduledAutocomplete.cancel(true));
      // debouncing the autocomplete trigger
      cancelCurrentJob();
      this.currentJob.set(cancellationToken);
    }
  }

  /** Asynchronously triggers auto-complete for the given editor and offset. */
  private CompletableFuture<Void> triggerAutocompleteAsync(
      @NotNull Project project,
      @NotNull Editor editor,
      int offset,
      @NotNull TextDocument textDocument,
      InlineCompletionTriggerKind triggerKind,
      CancellationToken cancellationToken) {
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
    CompletableFuture<InlineAutocompleteList> completions = server.autocompleteExecute(params);

    // Important: we have to `.cancel()` the original `CompletableFuture<T>` from lsp4j. As soon as
    // we use `thenAccept()` we get a new instance of `CompletableFuture<Void>` which does not
    // correctly propagate the cancelation to the agent.
    cancellationToken.onCancellationRequested(() -> completions.cancel(true));

    return completions
        .thenAccept(
            result ->
                processAutocompleteResult(
                    project, editor, offset, triggerKind, result, cancellationToken))
        .exceptionally(
            error -> {
              if (!(error instanceof CancellationException
                  || error instanceof CompletionException)) {
                logger.warn("failed autocomplete request " + params, error);
              }
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
      InlineAutocompleteList result,
      CancellationToken cancellationToken) {
    if (currentAutocompleteTelemetry != null) {
      currentAutocompleteTelemetry.markCompletionEvent(result.completionEvent);
    }

    if (Thread.interrupted() || cancellationToken.isCancelled()) {
      if (triggerKind.equals(InlineCompletionTriggerKind.INVOKE)) {
        logger.warn("autocomplete canceled");
      }
      return;
    }
    InlayModel inlayModel = editor.getInlayModel();
    Optional<InlineAutocompleteItem> maybeItem = result.items.stream().findFirst();
    if (maybeItem.isEmpty()) {
      if (triggerKind.equals(InlineCompletionTriggerKind.INVOKE)) {
        logger.warn("explicit autocomplete returned empty suggestions");
        // NOTE(olafur): it would be nice to give the user a visual hint when this happens.
        // We don't do anything now because it's unclear what would be the most idiomatic
        // IntelliJ API to use.
      }
      return;
    }
    final InlineAutocompleteItem item = maybeItem.get();
    ApplicationManager.getApplication()
        .invokeLater(
            () -> {
              if (cancellationToken.isCancelled()) {
                return;
              }
              cancellationToken.dispose();
              this.clearAutocompleteSuggestions(editor);

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
      InlineAutocompleteItem item,
      InlayModel inlayModel,
      InlineCompletionTriggerKind triggerKind) {
    TextRange range = CodyEditorUtil.getTextRange(editor.getDocument(), item.range);
    String originalText = editor.getDocument().getText(range);
    String insertTextFirstLine = item.insertText.lines().findFirst().orElse("");
    String multilineInsertText =
        item.insertText.lines().skip(1).collect(Collectors.joining(inferLineSeparator(editor)));

    // Run Myer's diff between the existing text in the document and the first line of the
    // `insertText` that is returned from the agent.
    // The diff algorithm returns a list of "deltas" that give us the minimal number of additions we
    // need to make to the document.
    Patch<String> patch = CodyAutocompleteManager.diff(originalText, insertTextFirstLine);
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
          new CodyAutocompleteSingleLineRenderer(
              text, item, editor, AutocompleteRendererType.INLINE));
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
          new CodyAutocompleteBlockElementRenderer(multilineInsertText, item, editor));
    }
  }

  private @NotNull String inferLineSeparator(@NotNull Editor editor) {
    VirtualFile virtualFile = FileDocumentManager.getInstance().getFile(editor.getDocument());

    if (virtualFile != null) {
      return Optional.ofNullable(virtualFile.getDetectedLineSeparator())
          .orElse(System.lineSeparator());
    } else {
      return System.lineSeparator();
    }
  }

  public static Patch<String> diff(String a, String b) {
    return DiffUtils.diff(characterList(a), characterList(b));
  }

  public static List<String> characterList(String value) {
    return value.chars().mapToObj(c -> String.valueOf((char) c)).collect(Collectors.toList());
  }

  // TODO: handle tabs in multiline autocomplete suggestions when we add them
  public static @NotNull InlineAutocompleteItem normalizeIndentation(
      @NotNull InlineAutocompleteItem item,
      @NotNull CommonCodeStyleSettings.IndentOptions indentOptions) {
    if (item.insertText.matches("^[\t ]*.+")) {
      String withoutLeadingWhitespace = item.insertText.stripLeading();
      String indentation =
          item.insertText.substring(
              0, item.insertText.length() - withoutLeadingWhitespace.length());
      String newIndentation = CodyEditorUtil.tabsToSpaces(indentation, indentOptions);
      String newInsertText = newIndentation + withoutLeadingWhitespace;
      int rangeDiff = item.insertText.length() - newInsertText.length();
      Range newRange =
          item.range.withEnd(item.range.end.withCharacter(item.range.end.character - rangeDiff));
      return item.withInsertText(newInsertText).withRange(newRange);
    } else return item;
  }

  private void cancelCurrentJob() {
    this.currentJob.get().abort();
    CodyAutocompleteStatusService.notifyApplication(CodyAutocompleteStatus.Ready);
  }
}
