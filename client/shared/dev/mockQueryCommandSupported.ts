/**
 * https://gitlab.com/gitlab-org/gitlab/-/issues/119194
 *
 * Monaco Editor requires functions that aren't available in JSDOM.
 */
document.queryCommandSupported = () => false
