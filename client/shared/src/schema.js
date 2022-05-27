"use strict";
exports.__esModule = true;
exports.SearchContextsOrderBy = exports.UserActivePeriod = exports.UserEvent = exports.AlertType = exports.RepositoryOrderBy = exports.OrganizationInvitationResponseType = exports.GitObjectType = exports.GitRefOrder = exports.GitRefType = exports.SymbolKind = exports.DiagnosticSeverity = exports.DiffHunkLineType = exports.ExternalServiceKind = exports.SearchPatternType = exports.SearchVersion = exports.EventSource = exports.NotebookBlockType = exports.NotebooksOrderBy = exports.TimeIntervalStepUnit = exports.SearchBasedSupportLevel = exports.PreciseSupportLevel = exports.InferedPreciseSupportLevel = exports.LSIFIndexState = exports.LSIFUploadState = exports.EventStatus = exports.MonitorEmailPriority = exports.BatchSpecState = exports.BulkOperationState = exports.BulkOperationType = exports.BatchChangeState = exports.BatchSpecWorkspaceState = exports.WorkspacesSortOrder = exports.BatchSpecWorkspaceResolutionState = exports.ChangesetSpecType = exports.ChangesetSpecOperation = exports.ChangesetState = exports.ChangesetCheckState = exports.ChangesetReviewState = exports.ChangesetExternalState = exports.ChangesetReconcilerState = exports.ChangesetPublicationState = exports.RepositoryPermission = void 0;
/**
 * Different repository permission levels.
 */
var RepositoryPermission;
(function (RepositoryPermission) {
    RepositoryPermission["READ"] = "READ";
})(RepositoryPermission = exports.RepositoryPermission || (exports.RepositoryPermission = {}));
/**
 * The publication state of a changeset on Sourcegraph
 */
var ChangesetPublicationState;
(function (ChangesetPublicationState) {
    /**
     * The changeset has not yet been created on the code host.
     */
    ChangesetPublicationState["UNPUBLISHED"] = "UNPUBLISHED";
    /**
     * The changeset has been created on the code host.
     */
    ChangesetPublicationState["PUBLISHED"] = "PUBLISHED";
})(ChangesetPublicationState = exports.ChangesetPublicationState || (exports.ChangesetPublicationState = {}));
/**
 * The reconciler state of a changeset on Sourcegraph
 */
var ChangesetReconcilerState;
(function (ChangesetReconcilerState) {
    /**
     * The changeset is scheduled, and will be enqueued when its turn comes in Sourcegraph's rollout window.
     */
    ChangesetReconcilerState["SCHEDULED"] = "SCHEDULED";
    /**
     * The changeset is enqueued for the reconciler to process it.
     */
    ChangesetReconcilerState["QUEUED"] = "QUEUED";
    /**
     * The changeset reconciler is currently computing the delta between the
     * If a delta exists, the reconciler tries to update the state of the
     * changeset on the code host and on Sourcegraph to the desired state.
     */
    ChangesetReconcilerState["PROCESSING"] = "PROCESSING";
    /**
     * The changeset reconciler ran into a problem while processing the
     * changeset and will retry it for a number of retries.
     */
    ChangesetReconcilerState["ERRORED"] = "ERRORED";
    /**
     * The changeset reconciler ran into a problem while processing the
     * changeset that can't be fixed by retrying.
     */
    ChangesetReconcilerState["FAILED"] = "FAILED";
    /**
     * The changeset is not enqueued for processing.
     */
    ChangesetReconcilerState["COMPLETED"] = "COMPLETED";
})(ChangesetReconcilerState = exports.ChangesetReconcilerState || (exports.ChangesetReconcilerState = {}));
/**
 * The state of a changeset on the code host on which it's hosted.
 */
var ChangesetExternalState;
(function (ChangesetExternalState) {
    ChangesetExternalState["DRAFT"] = "DRAFT";
    ChangesetExternalState["OPEN"] = "OPEN";
    ChangesetExternalState["CLOSED"] = "CLOSED";
    ChangesetExternalState["MERGED"] = "MERGED";
    ChangesetExternalState["DELETED"] = "DELETED";
})(ChangesetExternalState = exports.ChangesetExternalState || (exports.ChangesetExternalState = {}));
/**
 * The review state of a changeset.
 */
var ChangesetReviewState;
(function (ChangesetReviewState) {
    ChangesetReviewState["APPROVED"] = "APPROVED";
    ChangesetReviewState["CHANGES_REQUESTED"] = "CHANGES_REQUESTED";
    ChangesetReviewState["PENDING"] = "PENDING";
    ChangesetReviewState["COMMENTED"] = "COMMENTED";
    ChangesetReviewState["DISMISSED"] = "DISMISSED";
})(ChangesetReviewState = exports.ChangesetReviewState || (exports.ChangesetReviewState = {}));
/**
 * The state of checks (e.g., for continuous integration) on a changeset.
 */
var ChangesetCheckState;
(function (ChangesetCheckState) {
    ChangesetCheckState["PENDING"] = "PENDING";
    ChangesetCheckState["PASSED"] = "PASSED";
    ChangesetCheckState["FAILED"] = "FAILED";
})(ChangesetCheckState = exports.ChangesetCheckState || (exports.ChangesetCheckState = {}));
/**
 * The visual state a changeset is currently in.
 */
var ChangesetState;
(function (ChangesetState) {
    /**
     * The changeset has not been marked as to be published.
     */
    ChangesetState["UNPUBLISHED"] = "UNPUBLISHED";
    /**
     * The changeset reconciler ran into a problem while processing the
     * changeset that can't be fixed by retrying.
     */
    ChangesetState["FAILED"] = "FAILED";
    /**
     * The changeset reconciler ran into a problem while processing the
     * changeset and will retry it for a number of retries.
     */
    ChangesetState["RETRYING"] = "RETRYING";
    /**
     * The changeset is scheduled, and will be enqueued when its turn comes in Sourcegraph's rollout window.
     */
    ChangesetState["SCHEDULED"] = "SCHEDULED";
    /**
     * The changeset reconciler is currently computing the delta between the
     * If a delta exists, the reconciler tries to update the state of the
     * changeset on the code host and on Sourcegraph to the desired state.
     */
    ChangesetState["PROCESSING"] = "PROCESSING";
    /**
     * The changeset is published, not being reconciled and open on the code host.
     */
    ChangesetState["OPEN"] = "OPEN";
    /**
     * The changeset is published, not being reconciled and in draft state on the code host.
     */
    ChangesetState["DRAFT"] = "DRAFT";
    /**
     * The changeset is published, not being reconciled and closed on the code host.
     */
    ChangesetState["CLOSED"] = "CLOSED";
    /**
     * The changeset is published, not being reconciled and merged on the code host.
     */
    ChangesetState["MERGED"] = "MERGED";
    /**
     * The changeset is published, not being reconciled and has been deleted on the code host.
     */
    ChangesetState["DELETED"] = "DELETED";
})(ChangesetState = exports.ChangesetState || (exports.ChangesetState = {}));
/**
 * This enum declares all operations supported by the reconciler.
 */
var ChangesetSpecOperation;
(function (ChangesetSpecOperation) {
    /**
     * Push a new commit to the code host.
     */
    ChangesetSpecOperation["PUSH"] = "PUSH";
    /**
     * Update the existing changeset on the codehost. This is purely the changeset resource on the code host,
     * not the git commit. For updates to the commit, see 'PUSH'.
     */
    ChangesetSpecOperation["UPDATE"] = "UPDATE";
    /**
     * Move the existing changeset out of being a draft.
     */
    ChangesetSpecOperation["UNDRAFT"] = "UNDRAFT";
    /**
     * Publish a changeset to the codehost.
     */
    ChangesetSpecOperation["PUBLISH"] = "PUBLISH";
    /**
     * Publish a changeset to the codehost as a draft changeset. (Only on supported code hosts).
     */
    ChangesetSpecOperation["PUBLISH_DRAFT"] = "PUBLISH_DRAFT";
    /**
     * Sync the changeset with the current state on the codehost.
     */
    ChangesetSpecOperation["SYNC"] = "SYNC";
    /**
     * Import an existing changeset from the code host with the ExternalID from the spec.
     */
    ChangesetSpecOperation["IMPORT"] = "IMPORT";
    /**
     * Close the changeset on the codehost.
     */
    ChangesetSpecOperation["CLOSE"] = "CLOSE";
    /**
     * Reopen the changeset on the codehost.
     */
    ChangesetSpecOperation["REOPEN"] = "REOPEN";
    /**
     * Internal operation to get around slow code host updates.
     */
    ChangesetSpecOperation["SLEEP"] = "SLEEP";
    /**
     * The changeset is removed from some of the associated batch changes.
     */
    ChangesetSpecOperation["DETACH"] = "DETACH";
    /**
     * The changeset is kept in the batch change, but it's marked as archived.
     */
    ChangesetSpecOperation["ARCHIVE"] = "ARCHIVE";
})(ChangesetSpecOperation = exports.ChangesetSpecOperation || (exports.ChangesetSpecOperation = {}));
/**
 * The type of the changeset spec.
 */
var ChangesetSpecType;
(function (ChangesetSpecType) {
    /**
     * References an existing changeset on a code host to be imported.
     */
    ChangesetSpecType["EXISTING"] = "EXISTING";
    /**
     * References a branch and a patch to be applied to create the changeset from.
     */
    ChangesetSpecType["BRANCH"] = "BRANCH";
})(ChangesetSpecType = exports.ChangesetSpecType || (exports.ChangesetSpecType = {}));
/**
 * State of the workspace resolution.
 */
var BatchSpecWorkspaceResolutionState;
(function (BatchSpecWorkspaceResolutionState) {
    /**
     * Not yet started resolving. Will be picked up by a worker eventually.
     */
    BatchSpecWorkspaceResolutionState["QUEUED"] = "QUEUED";
    /**
     * Currently resolving workspaces.
     */
    BatchSpecWorkspaceResolutionState["PROCESSING"] = "PROCESSING";
    /**
     * An error occured while resolving workspaces. Will be retried eventually.
     */
    BatchSpecWorkspaceResolutionState["ERRORED"] = "ERRORED";
    /**
     * A fatal error occured while resolving workspaces. No retries will be made.
     */
    BatchSpecWorkspaceResolutionState["FAILED"] = "FAILED";
    /**
     * Resolving workspaces finished successfully.
     */
    BatchSpecWorkspaceResolutionState["COMPLETED"] = "COMPLETED";
})(BatchSpecWorkspaceResolutionState = exports.BatchSpecWorkspaceResolutionState || (exports.BatchSpecWorkspaceResolutionState = {}));
/**
 * Possible sort orderings for a workspace connection.
 */
var WorkspacesSortOrder;
(function (WorkspacesSortOrder) {
    /**
     * Sort by repository name in ascending order.
     */
    WorkspacesSortOrder["REPO_NAME_ASC"] = "REPO_NAME_ASC";
    /**
     * Sort by repository name in descending order.
     */
    WorkspacesSortOrder["REPO_NAME_DESC"] = "REPO_NAME_DESC";
})(WorkspacesSortOrder = exports.WorkspacesSortOrder || (exports.WorkspacesSortOrder = {}));
/**
 * The states a workspace can be in.
 */
var BatchSpecWorkspaceState;
(function (BatchSpecWorkspaceState) {
    /**
     * The workspace will not be enqueued for execution, because either the
     * workspace is unsupported/ignored or has 0 steps to execute.
     */
    BatchSpecWorkspaceState["SKIPPED"] = "SKIPPED";
    /**
     * The workspace is not yet enqueued for execution.
     */
    BatchSpecWorkspaceState["PENDING"] = "PENDING";
    /**
     * Not yet started executing. Will be picked up by a worker eventually.
     */
    BatchSpecWorkspaceState["QUEUED"] = "QUEUED";
    /**
     * Currently executing on the workspace.
     */
    BatchSpecWorkspaceState["PROCESSING"] = "PROCESSING";
    /**
     * A fatal error occured while executing. No retries will be made.
     */
    BatchSpecWorkspaceState["FAILED"] = "FAILED";
    /**
     * Execution finished successfully.
     */
    BatchSpecWorkspaceState["COMPLETED"] = "COMPLETED";
    /**
     * Execution is being canceled. This is an async process.
     */
    BatchSpecWorkspaceState["CANCELING"] = "CANCELING";
    /**
     * Execution has been canceled.
     */
    BatchSpecWorkspaceState["CANCELED"] = "CANCELED";
})(BatchSpecWorkspaceState = exports.BatchSpecWorkspaceState || (exports.BatchSpecWorkspaceState = {}));
/**
 * The state of the batch change.
 */
var BatchChangeState;
(function (BatchChangeState) {
    BatchChangeState["OPEN"] = "OPEN";
    BatchChangeState["CLOSED"] = "CLOSED";
    BatchChangeState["DRAFT"] = "DRAFT";
})(BatchChangeState = exports.BatchChangeState || (exports.BatchChangeState = {}));
/**
 * The available types of jobs that can be run over a batch change.
 */
var BulkOperationType;
(function (BulkOperationType) {
    /**
     * Bulk post comments over all involved changesets.
     */
    BulkOperationType["COMMENT"] = "COMMENT";
    /**
     * Bulk detach changesets from a batch change.
     */
    BulkOperationType["DETACH"] = "DETACH";
    /**
     * Bulk reenqueue failed changesets.
     */
    BulkOperationType["REENQUEUE"] = "REENQUEUE";
    /**
     * Bulk merge changesets.
     */
    BulkOperationType["MERGE"] = "MERGE";
    /**
     * Bulk close changesets.
     */
    BulkOperationType["CLOSE"] = "CLOSE";
    /**
     * Bulk publish changesets.
     */
    BulkOperationType["PUBLISH"] = "PUBLISH";
})(BulkOperationType = exports.BulkOperationType || (exports.BulkOperationType = {}));
/**
 * All valid states a bulk operation can be in.
 */
var BulkOperationState;
(function (BulkOperationState) {
    /**
     * The bulk operation is still processing on some changesets.
     */
    BulkOperationState["PROCESSING"] = "PROCESSING";
    /**
     * No operations are still running and all of them finished without error.
     */
    BulkOperationState["COMPLETED"] = "COMPLETED";
    /**
     * No operations are still running and at least one of them finished with an error.
     */
    BulkOperationState["FAILED"] = "FAILED";
})(BulkOperationState = exports.BulkOperationState || (exports.BulkOperationState = {}));
/**
 * The possible states of a batch spec.
 */
var BatchSpecState;
(function (BatchSpecState) {
    /**
     * The spec is not yet enqueued for processing.
     */
    BatchSpecState["PENDING"] = "PENDING";
    /**
     * This spec is being processed.
     */
    BatchSpecState["PROCESSING"] = "PROCESSING";
    /**
     * This spec failed to be processed.
     */
    BatchSpecState["FAILED"] = "FAILED";
    /**
     * This spec was processed successfully.
     */
    BatchSpecState["COMPLETED"] = "COMPLETED";
    /**
     * This spec is queued to be processed.
     */
    BatchSpecState["QUEUED"] = "QUEUED";
    /**
     * The execution is being canceled.
     */
    BatchSpecState["CANCELING"] = "CANCELING";
    /**
     * The execution has been canceled.
     */
    BatchSpecState["CANCELED"] = "CANCELED";
})(BatchSpecState = exports.BatchSpecState || (exports.BatchSpecState = {}));
/**
 * The priority of an email action.
 */
var MonitorEmailPriority;
(function (MonitorEmailPriority) {
    MonitorEmailPriority["NORMAL"] = "NORMAL";
    MonitorEmailPriority["CRITICAL"] = "CRITICAL";
})(MonitorEmailPriority = exports.MonitorEmailPriority || (exports.MonitorEmailPriority = {}));
/**
 * Supported status of monitor events.
 */
var EventStatus;
(function (EventStatus) {
    EventStatus["PENDING"] = "PENDING";
    EventStatus["SUCCESS"] = "SUCCESS";
    EventStatus["ERROR"] = "ERROR";
})(EventStatus = exports.EventStatus || (exports.EventStatus = {}));
/**
 * The state an LSIF upload can be in.
 */
var LSIFUploadState;
(function (LSIFUploadState) {
    /**
     * This upload is being processed.
     */
    LSIFUploadState["PROCESSING"] = "PROCESSING";
    /**
     * This upload failed to be processed.
     */
    LSIFUploadState["ERRORED"] = "ERRORED";
    /**
     * This upload was processed successfully.
     */
    LSIFUploadState["COMPLETED"] = "COMPLETED";
    /**
     * This upload is queued to be processed later.
     */
    LSIFUploadState["QUEUED"] = "QUEUED";
    /**
     * This upload is currently being transferred to Sourcegraph.
     */
    LSIFUploadState["UPLOADING"] = "UPLOADING";
    /**
     * This upload is queued for deletion. This upload was previously in the
     * COMPLETED state and evicted, replaced by a newer upload, or deleted by
     * a user. This upload is able to answer code intelligence queries until
     * the commit graph of the upload's repository is next calculated, at which
     * point the upload will become unreachable.
     */
    LSIFUploadState["DELETING"] = "DELETING";
})(LSIFUploadState = exports.LSIFUploadState || (exports.LSIFUploadState = {}));
/**
 * The state an LSIF index can be in.
 */
var LSIFIndexState;
(function (LSIFIndexState) {
    /**
     * This index is being processed.
     */
    LSIFIndexState["PROCESSING"] = "PROCESSING";
    /**
     * This index failed to be processed.
     */
    LSIFIndexState["ERRORED"] = "ERRORED";
    /**
     * This index was processed successfully.
     */
    LSIFIndexState["COMPLETED"] = "COMPLETED";
    /**
     * This index is queued to be processed later.
     */
    LSIFIndexState["QUEUED"] = "QUEUED";
})(LSIFIndexState = exports.LSIFIndexState || (exports.LSIFIndexState = {}));
/**
 * Denotes the confidence in the correctness of the proposed index target.
 */
var InferedPreciseSupportLevel;
(function (InferedPreciseSupportLevel) {
    /**
     * The language is known to have an LSIF indexer associated with it
     * but this may not be the directory from which it should be invoked.
     * Relevant build tool configuration may be available at a parent directory.
     */
    InferedPreciseSupportLevel["LANGUAGE_SUPPORTED"] = "LANGUAGE_SUPPORTED";
    /**
     * Relevant build tool configuration files were located that indicate
     * a good possibility of this directory being where an LSIF indexer
     * could be invoked, however we have or can not infer a potentially complete
     * auto indexing job configuration.
     */
    InferedPreciseSupportLevel["PROJECT_STRUCTURE_SUPPORTED"] = "PROJECT_STRUCTURE_SUPPORTED";
    /**
     * An auto-indexing job configuration was able to be infered for this
     * directory that has a high likelyhood of being complete enough to result
     * in an LSIF index.
     */
    InferedPreciseSupportLevel["INDEX_JOB_INFERED"] = "INDEX_JOB_INFERED";
})(InferedPreciseSupportLevel = exports.InferedPreciseSupportLevel || (exports.InferedPreciseSupportLevel = {}));
/**
 * Ownership level of the recommended precise code-intel indexer.
 */
var PreciseSupportLevel;
(function (PreciseSupportLevel) {
    /**
     * When there is no known indexer.
     */
    PreciseSupportLevel["UNKNOWN"] = "UNKNOWN";
    /**
     * When the recommended indexer is maintained by us.
     */
    PreciseSupportLevel["NATIVE"] = "NATIVE";
    /**
     * When the recommended indexer is maintained by a third-party
     * but is recommended over a native indexer, where one exists.
     */
    PreciseSupportLevel["THIRD_PARTY"] = "THIRD_PARTY";
})(PreciseSupportLevel = exports.PreciseSupportLevel || (exports.PreciseSupportLevel = {}));
/**
 * Tiered list of types of search-based support for a language. This may be expanded as different
 * indexing methods are introduced.
 */
var SearchBasedSupportLevel;
(function (SearchBasedSupportLevel) {
    /**
     * The language has no configured search-based code-intel support.
     */
    SearchBasedSupportLevel["UNSUPPORTED"] = "UNSUPPORTED";
    /**
     * Universal-ctags is used for indexing this language.
     */
    SearchBasedSupportLevel["BASIC"] = "BASIC";
})(SearchBasedSupportLevel = exports.SearchBasedSupportLevel || (exports.SearchBasedSupportLevel = {}));
/**
 * Time interval units.
 */
var TimeIntervalStepUnit;
(function (TimeIntervalStepUnit) {
    TimeIntervalStepUnit["HOUR"] = "HOUR";
    TimeIntervalStepUnit["DAY"] = "DAY";
    TimeIntervalStepUnit["WEEK"] = "WEEK";
    TimeIntervalStepUnit["MONTH"] = "MONTH";
    TimeIntervalStepUnit["YEAR"] = "YEAR";
})(TimeIntervalStepUnit = exports.TimeIntervalStepUnit || (exports.TimeIntervalStepUnit = {}));
/**
 * NotebooksOrderBy enumerates the ways notebooks can be ordered.
 */
var NotebooksOrderBy;
(function (NotebooksOrderBy) {
    NotebooksOrderBy["NOTEBOOK_UPDATED_AT"] = "NOTEBOOK_UPDATED_AT";
    NotebooksOrderBy["NOTEBOOK_CREATED_AT"] = "NOTEBOOK_CREATED_AT";
    NotebooksOrderBy["NOTEBOOK_STAR_COUNT"] = "NOTEBOOK_STAR_COUNT";
})(NotebooksOrderBy = exports.NotebooksOrderBy || (exports.NotebooksOrderBy = {}));
/**
 * Enum of possible block types.
 */
var NotebookBlockType;
(function (NotebookBlockType) {
    NotebookBlockType["MARKDOWN"] = "MARKDOWN";
    NotebookBlockType["QUERY"] = "QUERY";
    NotebookBlockType["FILE"] = "FILE";
    NotebookBlockType["SYMBOL"] = "SYMBOL";
    NotebookBlockType["COMPUTE"] = "COMPUTE";
})(NotebookBlockType = exports.NotebookBlockType || (exports.NotebookBlockType = {}));
/**
 * The product sources where events can come from.
 */
var EventSource;
(function (EventSource) {
    EventSource["WEB"] = "WEB";
    EventSource["CODEHOSTINTEGRATION"] = "CODEHOSTINTEGRATION";
    EventSource["BACKEND"] = "BACKEND";
    EventSource["STATICWEB"] = "STATICWEB";
    EventSource["IDEEXTENSION"] = "IDEEXTENSION";
})(EventSource = exports.EventSource || (exports.EventSource = {}));
/**
 * The version of the search syntax.
 */
var SearchVersion;
(function (SearchVersion) {
    /**
     * Search syntax that defaults to regexp search.
     */
    SearchVersion["V1"] = "V1";
    /**
     * Search syntax that defaults to literal search.
     */
    SearchVersion["V2"] = "V2";
})(SearchVersion = exports.SearchVersion || (exports.SearchVersion = {}));
/**
 * The search pattern type.
 */
var SearchPatternType;
(function (SearchPatternType) {
    SearchPatternType["literal"] = "literal";
    SearchPatternType["regexp"] = "regexp";
    SearchPatternType["structural"] = "structural";
})(SearchPatternType = exports.SearchPatternType || (exports.SearchPatternType = {}));
/**
 * A specific kind of external service.
 */
var ExternalServiceKind;
(function (ExternalServiceKind) {
    ExternalServiceKind["AWSCODECOMMIT"] = "AWSCODECOMMIT";
    ExternalServiceKind["BITBUCKETCLOUD"] = "BITBUCKETCLOUD";
    ExternalServiceKind["BITBUCKETSERVER"] = "BITBUCKETSERVER";
    ExternalServiceKind["GERRIT"] = "GERRIT";
    ExternalServiceKind["GITHUB"] = "GITHUB";
    ExternalServiceKind["GITLAB"] = "GITLAB";
    ExternalServiceKind["GITOLITE"] = "GITOLITE";
    ExternalServiceKind["GOMODULES"] = "GOMODULES";
    ExternalServiceKind["JVMPACKAGES"] = "JVMPACKAGES";
    ExternalServiceKind["NPMPACKAGES"] = "NPMPACKAGES";
    ExternalServiceKind["OTHER"] = "OTHER";
    ExternalServiceKind["PAGURE"] = "PAGURE";
    ExternalServiceKind["PERFORCE"] = "PERFORCE";
    ExternalServiceKind["PHABRICATOR"] = "PHABRICATOR";
    ExternalServiceKind["PYTHONPACKAGES"] = "PYTHONPACKAGES";
})(ExternalServiceKind = exports.ExternalServiceKind || (exports.ExternalServiceKind = {}));
/**
 * The type of content in a hunk line.
 */
var DiffHunkLineType;
(function (DiffHunkLineType) {
    /**
     * Added line.
     */
    DiffHunkLineType["ADDED"] = "ADDED";
    /**
     * Unchanged line.
     */
    DiffHunkLineType["UNCHANGED"] = "UNCHANGED";
    /**
     * Deleted line.
     */
    DiffHunkLineType["DELETED"] = "DELETED";
})(DiffHunkLineType = exports.DiffHunkLineType || (exports.DiffHunkLineType = {}));
/**
 * Represents the severity level of a diagnostic.
 */
var DiagnosticSeverity;
(function (DiagnosticSeverity) {
    DiagnosticSeverity["ERROR"] = "ERROR";
    DiagnosticSeverity["WARNING"] = "WARNING";
    DiagnosticSeverity["INFORMATION"] = "INFORMATION";
    DiagnosticSeverity["HINT"] = "HINT";
})(DiagnosticSeverity = exports.DiagnosticSeverity || (exports.DiagnosticSeverity = {}));
/**
 * All possible kinds of symbols. This set matches that of the Language Server Protocol
 * (https://microsoft.github.io/language-server-protocol/specification#workspace_symbol).
 */
var SymbolKind;
(function (SymbolKind) {
    SymbolKind["UNKNOWN"] = "UNKNOWN";
    SymbolKind["FILE"] = "FILE";
    SymbolKind["MODULE"] = "MODULE";
    SymbolKind["NAMESPACE"] = "NAMESPACE";
    SymbolKind["PACKAGE"] = "PACKAGE";
    SymbolKind["CLASS"] = "CLASS";
    SymbolKind["METHOD"] = "METHOD";
    SymbolKind["PROPERTY"] = "PROPERTY";
    SymbolKind["FIELD"] = "FIELD";
    SymbolKind["CONSTRUCTOR"] = "CONSTRUCTOR";
    SymbolKind["ENUM"] = "ENUM";
    SymbolKind["INTERFACE"] = "INTERFACE";
    SymbolKind["FUNCTION"] = "FUNCTION";
    SymbolKind["VARIABLE"] = "VARIABLE";
    SymbolKind["CONSTANT"] = "CONSTANT";
    SymbolKind["STRING"] = "STRING";
    SymbolKind["NUMBER"] = "NUMBER";
    SymbolKind["BOOLEAN"] = "BOOLEAN";
    SymbolKind["ARRAY"] = "ARRAY";
    SymbolKind["OBJECT"] = "OBJECT";
    SymbolKind["KEY"] = "KEY";
    SymbolKind["NULL"] = "NULL";
    SymbolKind["ENUMMEMBER"] = "ENUMMEMBER";
    SymbolKind["STRUCT"] = "STRUCT";
    SymbolKind["EVENT"] = "EVENT";
    SymbolKind["OPERATOR"] = "OPERATOR";
    SymbolKind["TYPEPARAMETER"] = "TYPEPARAMETER";
})(SymbolKind = exports.SymbolKind || (exports.SymbolKind = {}));
/**
 * All possible types of Git refs.
 */
var GitRefType;
(function (GitRefType) {
    /**
     * A Git branch (in refs/heads/).
     */
    GitRefType["GIT_BRANCH"] = "GIT_BRANCH";
    /**
     * A Git tag (in refs/tags/).
     */
    GitRefType["GIT_TAG"] = "GIT_TAG";
    /**
     * A Git ref that is neither a branch nor tag.
     */
    GitRefType["GIT_REF_OTHER"] = "GIT_REF_OTHER";
})(GitRefType = exports.GitRefType || (exports.GitRefType = {}));
/**
 * Ordering options for Git refs.
 */
var GitRefOrder;
(function (GitRefOrder) {
    /**
     * By the authored or committed at date, whichever is more recent.
     */
    GitRefOrder["AUTHORED_OR_COMMITTED_AT"] = "AUTHORED_OR_COMMITTED_AT";
})(GitRefOrder = exports.GitRefOrder || (exports.GitRefOrder = {}));
/**
 * All possible types of Git objects.
 */
var GitObjectType;
(function (GitObjectType) {
    /**
     * A Git commit object.
     */
    GitObjectType["GIT_COMMIT"] = "GIT_COMMIT";
    /**
     * A Git tag object.
     */
    GitObjectType["GIT_TAG"] = "GIT_TAG";
    /**
     * A Git tree object.
     */
    GitObjectType["GIT_TREE"] = "GIT_TREE";
    /**
     * A Git blob object.
     */
    GitObjectType["GIT_BLOB"] = "GIT_BLOB";
    /**
     * A Git object of unknown type.
     */
    GitObjectType["GIT_UNKNOWN"] = "GIT_UNKNOWN";
})(GitObjectType = exports.GitObjectType || (exports.GitObjectType = {}));
/**
 * The recipient's possible responses to an invitation to join an organization as a member.
 */
var OrganizationInvitationResponseType;
(function (OrganizationInvitationResponseType) {
    /**
     * The invitation was accepted by the recipient.
     */
    OrganizationInvitationResponseType["ACCEPT"] = "ACCEPT";
    /**
     * The invitation was rejected by the recipient.
     */
    OrganizationInvitationResponseType["REJECT"] = "REJECT";
})(OrganizationInvitationResponseType = exports.OrganizationInvitationResponseType || (exports.OrganizationInvitationResponseType = {}));
/**
 * RepositoryOrderBy enumerates the ways a repositories list can be ordered.
 */
var RepositoryOrderBy;
(function (RepositoryOrderBy) {
    RepositoryOrderBy["REPOSITORY_NAME"] = "REPOSITORY_NAME";
    RepositoryOrderBy["REPO_CREATED_AT"] = "REPO_CREATED_AT";
    /**
     * deprecated (use the equivalent REPOSITORY_CREATED_AT)
     */
    RepositoryOrderBy["REPOSITORY_CREATED_AT"] = "REPOSITORY_CREATED_AT";
})(RepositoryOrderBy = exports.RepositoryOrderBy || (exports.RepositoryOrderBy = {}));
/**
 * The possible types of alerts (Alert.type values).
 */
var AlertType;
(function (AlertType) {
    AlertType["INFO"] = "INFO";
    AlertType["WARNING"] = "WARNING";
    AlertType["ERROR"] = "ERROR";
})(AlertType = exports.AlertType || (exports.AlertType = {}));
/**
 * A user event.
 */
var UserEvent;
(function (UserEvent) {
    UserEvent["PAGEVIEW"] = "PAGEVIEW";
    UserEvent["SEARCHQUERY"] = "SEARCHQUERY";
    UserEvent["CODEINTEL"] = "CODEINTEL";
    UserEvent["CODEINTELREFS"] = "CODEINTELREFS";
    UserEvent["CODEINTELINTEGRATION"] = "CODEINTELINTEGRATION";
    UserEvent["CODEINTELINTEGRATIONREFS"] = "CODEINTELINTEGRATIONREFS";
    /**
     * Product stages
     */
    UserEvent["STAGEMANAGE"] = "STAGEMANAGE";
    UserEvent["STAGEPLAN"] = "STAGEPLAN";
    UserEvent["STAGECODE"] = "STAGECODE";
    UserEvent["STAGEREVIEW"] = "STAGEREVIEW";
    UserEvent["STAGEVERIFY"] = "STAGEVERIFY";
    UserEvent["STAGEPACKAGE"] = "STAGEPACKAGE";
    UserEvent["STAGEDEPLOY"] = "STAGEDEPLOY";
    UserEvent["STAGECONFIGURE"] = "STAGECONFIGURE";
    UserEvent["STAGEMONITOR"] = "STAGEMONITOR";
    UserEvent["STAGESECURE"] = "STAGESECURE";
    UserEvent["STAGEAUTOMATE"] = "STAGEAUTOMATE";
})(UserEvent = exports.UserEvent || (exports.UserEvent = {}));
/**
 * A period of time in which a set of users have been active.
 */
var UserActivePeriod;
(function (UserActivePeriod) {
    /**
     * Since today at 00:00 UTC.
     */
    UserActivePeriod["TODAY"] = "TODAY";
    /**
     * Since the latest Monday at 00:00 UTC.
     */
    UserActivePeriod["THIS_WEEK"] = "THIS_WEEK";
    /**
     * Since the first day of the current month at 00:00 UTC.
     */
    UserActivePeriod["THIS_MONTH"] = "THIS_MONTH";
    /**
     * All time.
     */
    UserActivePeriod["ALL_TIME"] = "ALL_TIME";
})(UserActivePeriod = exports.UserActivePeriod || (exports.UserActivePeriod = {}));
/**
 * SearchContextsOrderBy enumerates the ways a search contexts list can be ordered.
 */
var SearchContextsOrderBy;
(function (SearchContextsOrderBy) {
    SearchContextsOrderBy["SEARCH_CONTEXT_SPEC"] = "SEARCH_CONTEXT_SPEC";
    SearchContextsOrderBy["SEARCH_CONTEXT_UPDATED_AT"] = "SEARCH_CONTEXT_UPDATED_AT";
})(SearchContextsOrderBy = exports.SearchContextsOrderBy || (exports.SearchContextsOrderBy = {}));
