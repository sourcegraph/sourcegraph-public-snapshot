var $ = require("jquery");
var router = require("../routing/router");
var globals = require("../globals");

/**
 * @description The running XHR request for a popup.
 * @type {jqXHR}
 */
var popupXhr = null;

/**
 * @description Popup cache. Maps URLs to previously requested data.
 */
var popupCache = {};

/**
 * @description The running XHR request for a popover.
 * @type {jqXHR}
 */
var popoverXhr = null;

/**
 * @description The running XHR request for a context menu.
 * @type {jqXHR}
 */
var defListXhr = null;

/**
 * @description Popover cache. Maps URLs to popover data.
 */
var popoverCache = {};

/**
 * @description The running XHR request for an example.
 * @type {jqXHR}
 */
var exampleXhr = null;

module.exports = {
	/**
	 * @description Fetches file contents and information for the passed file URL which
	 * may be a tree entry or a definition URL, in which case definition / popup data
	 * is also returned along with the file containing the definition.
	 * @param {string} url - Tree entry or definition URL.
	 * @param {number=} start - Optional start line.
	 * @param {number=} end - Optional end line.
	 * @returns {jQuery.jqXHR} - Promise
	 */
	fetchFile(url, start, end) {
		var opt = "";
		if (start) {
			opt = `?StartLine=${start}${end ? `&EndLine=${end}` : ""}`;
		}
		return $.ajax({
			url: `/ui${url}${opt}`,
			contentType: "application/json",
		}).then(module.exports.receivedFile);
	},

	/**
	 * @description Sends the request to create a new changeset and returns
	 * the promise. A succesful promise will return the full changeset model.
	 * @param {string} repo - Repository URL.
	 * @param {object} changeSet - Changeset to open.
	 * @returns {jQuery.jqXHR} - Promise.
	 */
	createChangeset(repo, changeSet) {
		var createUrl = `${router.repoURL(repo)}/.changes/create`;

		return $.ajax({
			method: "POST",
			headers: {
				"X-CSRF-Token": globals.CsrfToken,
			},
			url: createUrl,
			data: JSON.stringify(changeSet),
		}).then(data => {
			if (data.hasOwnProperty("Error")) {
				return $.Deferred().reject(data.Error);
			}
			return data;
		});
	},

	/**
	 * @description Triggers the server action that a file was received.
	 * @param {object} data - File data returned from the server or cache.
	 * @returns {jQuery.jqXHR} - promise
	 */
	receivedFile(data) {
		if (data.hasOwnProperty("RedirectTo")) {
			return $.Deferred().reject({RedirectTo: data.RedirectTo});
		}
		if (data.Definition) {
			popupCache[data.Definition.URL] = data.Definition;
		}

		return $.Deferred().resolve(data);
	},

	/**
	 * @description Fetches a list of multiple definitions in one go.
	 * @param {Array<string>} urls - List of DefKey's to return definitions for.
	 * @returns {Array<DefCommon>} - List of defs.
	 */
	fetchDefinitionList(urls) {
		if (defListXhr) defListXhr.abort();

		var fromCache = [];
		var fromServer = urls.filter(url => {
			if (popupCache.hasOwnProperty(url)) {
				fromCache.push(popupCache[url]);
				return false;
			}
			return true;
		});

		if (fromServer.length === 0) {
			return $.Deferred().resolve(fromCache);
		}

		var onDone = data => {
			if (!Array.isArray(data.Defs)) {
				return $.Deferred().reject({Error: "Invalid response"});
			}
			data.Defs.forEach(def => {
				if (!popupCache.hasOwnProperty(def.URL)) {
					popupCache[def.URL] = def;
				}
			});
			return $.Deferred().resolve(fromCache.concat(data.Defs));
		};

		var onError = (data, status) => {
			if (status !== "abort") {
				return $.Deferred().reject();
			}
		};

		defListXhr = $.ajax({
			url: `/ui/.defs?key=${fromServer.map(encodeURIComponent).join("&key=")}`,
		});

		return defListXhr.then(onDone, onError).always(() => { defListXhr = null; });
	},

	/**
	 * @description Fetches the definition object / popup data for the definition at the
	 * given URL.
	 * @param {string} url - Definition URL.
	 * @returns {jQuery.jqXHR} - Promise
	 */
	fetchPopup(url) {
		if (popupCache[url]) {
			return $.Deferred().resolve(popupCache[url]);
		}

		if (defListXhr) defListXhr.abort();
		if (popupXhr) popupXhr.abort();
		if (popoverXhr) popoverXhr.abort();
		if (exampleXhr) exampleXhr.abort();

		popupXhr = $.ajax({
			url: `/ui${url}`,
			headers: {
				"X-Definition-Data-Only": "yes",
			},
		});

		var receivedPopup = function(defUrl, data) {
			if (data.hasOwnProperty("Error")) {
				return $.Deferred().reject(data.Error);
			}
			popupXhr = null;
			popupCache[defUrl] = data;
			return data;
		};

		return popupXhr.then(receivedPopup.bind(this, url));
	},

	/**
	 * @description Fetches popover data for the definition at the given URL.
	 * @param {string} url - Definition URL.
	 * @returns {jQuery.jqXHR} - Promise
	 */
	fetchPopover(url) {
		if (popoverCache[url]) {
			return $.Deferred().resolve(popoverCache[url]);
		}

		module.exports.abortPopoverXhr();

		popoverXhr = $.ajax({url: `${url}/.popover`});

		var receivedPopover = function(data) {
			popoverXhr = null;
			var body = {__html: data};
			popoverCache[url] = body;
			return body;
		};

		return popoverXhr.then(receivedPopover);
	},

	/**
	 * @description Aborts any running popover requests. Useful when user mouses over and out
	 * of token to reduce unnecessary requests.
	 * @returns {void}
	 */
	abortPopoverXhr() {
		if (popoverXhr) popoverXhr.abort();
	},

	/**
	 * @description Fetches usage examples for the passed definition URL, having the given
	 * page offset.
	 * @param {string} url - Definition URL.
	 * @param {number} page - Page offset.
	 * @param {string} fallbackRepoURI - Try find examples in this repo if we can't find any.
	 * @returns {jQuery.jqXHR} - Promise
	 */
	fetchExample(url, page, fallbackRepoURI) {
		var data = {
			TokenizedSource: true,
			PerPage: 1,
			Page: page,
			FallbackRepoURI: fallbackRepoURI,
		};
		var opts = {
			url: `/ui${url}/.examples`,
			data: data,
			dataType: "json",
		};

		if (exampleXhr) exampleXhr.abort();

		exampleXhr = $.ajax(opts);

		var receivedExample = function(data2) {
			return {
				example: data2,
				page: page,
			};
		};

		return exampleXhr.then(receivedExample).always(() => exampleXhr = null);
	},

	/**
	 * @description Submits the review and returns a promise object.
	 * @param {string} repo - Repo path.
	 * @param {number} changesetId - ID of changeset.
	 * @param {string} body - Text body of review.
	 * @param {Array<Object>} drafts - Array of comment model attributes.
	 * @returns {jQuery.jqXHR} - Promise.
	 */
	submitReview(repo, changesetId, body, drafts) {
		var url = `${router.changesetURL(repo, changesetId)}/submit-review`;

		return $.ajax({
			url: url,
			method: "POST",
			headers: {
				"X-CSRF-Token": globals.CsrfToken,
			},
			data: JSON.stringify({
				Body: body,
				Comments: drafts,
				CreatedAt: new Date(),
			}),
		}).then(data => {
			if (data.hasOwnProperty("Error")) {
				return $.Deferred().reject(data.Error);
			}
			return data;
		});
	},

	/**
	 * @description Updates a changesets status and returns the promise. When
	 * complete, the promise returns the event that occurred, if any.
	 * @param {string} repo - Repository URL.
	 * @param {number} changesetId - Changeset ID.
	 * @param {sourcegraph.ChangesetUpdateOp} status - Updated status.
	 * @returns {jQuery.jqXHR} - Promise.
	 */
	updateChangesetStatus(repo, changesetId, status) {
		var url = `${router.changesetURL(repo, changesetId)}/update`;

		return $.ajax({
			url: url,
			method: "POST",
			headers: {
				"X-CSRF-Token": globals.CsrfToken,
			},
			data: JSON.stringify(status),
		}).then(data => {
			if (data.hasOwnProperty("Error")) {
				return $.Deferred().reject(data.Error);
			}
			return data;
		});
	},

	submitDiscussionComment(defKey, id, body) {
		var url = `/ui${router.discussionCreateCommentURL(defKey, id)}`;
		var desc = {
			Body: body,
		};
		return $.ajax({
			url: url,
			method: "POST",
			data: JSON.stringify(desc),
		}).then(data => {
			if (data.hasOwnProperty("Error")) {
				return $.Deferred().reject(data.Error);
			}
			return data;
		});
	},

	submitDiscussion(defKey, title, body) {
		var url = `/ui${router.discussionCreateURL(defKey)}`;
		var desc = {
			Title: title,
			Description: body,
		};
		return $.ajax({
			url: url,
			method: "POST",
			data: JSON.stringify(desc),
		}).then(data => {
			if (data.hasOwnProperty("Error")) {
				return $.Deferred().reject(data.Error);
			}
			return data;
		});
	},

	fetchDiscussion(defKey, dsc) {
		return $.ajax({url: `/ui/${router.discussionURL(defKey, dsc.ID)}`}).then(data => {
			if (data.hasOwnProperty("Error")) {
				return $.Deferred().reject(data.Error);
			}
			return data;
		});
	},

	fetchTopDiscussions(defKey) {
		if (!globals.Features.Discussions) {
			return $.Deferred().resolve({Discussions: []});
		}
		return $.ajax({url: `/ui${router.discussionListURL(defKey, "Top")}`}).then(data => {
			if (data.hasOwnProperty("Error")) {
				return $.Deferred().reject(data.Error);
			}
			return data;
		});
	},

	fetchDiscussionList(defKey) {
		return $.ajax({url: `/ui${router.discussionListURL(defKey, "Date")}`});
	},

	fetchAllDiscussions(repo) {
		var url = `/ui${router.repoURL(repo)}/.discussions?order=Date`;
		return $.ajax({url: url}).then(data => {
			if (data.hasOwnProperty("Error")) {
				return $.Deferred().reject(data.Error);
			}
			return data;
		});
	},
};
