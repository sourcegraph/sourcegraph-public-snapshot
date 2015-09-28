var $ = require("jquery");

var defaultTimeout = 3000;

/*
 * @description Displays a notification that timeouts after a certain amount of time.
 * @param {string} cls - The class to assign to the notification's DOM node
 * @param {string} message - The message to display
 * @param {number=} timeout - Optional parameter. Timeout in milliseconds. Defaults to 3s.
 */
function displayNotification(cls, message, timeout) {
	var iconCls = {
		warning: "fa-warning",
		success: "fa-check-circle",
		error: "fa-exclamation-triangle",
		info: "fa-info-circle",
	}[cls];

	var $el = $(
		"<div class=\"alert-notify "+cls+"\">" +
			"<a class=\"close\">×</a>" +
			"<i class=\"fa "+iconCls+"\"></i>" +
			message +
		"</div>"
	);

	var remove = () => $el.fadeOut(250, "linear", () => $el.remove());

	$("body").append($el);

	$el
		.hide()
		.css({
			"top": findOffset($el),
			"margin-left": -($el.outerWidth() / 2),
		})
		.fadeIn(450, "linear")
		.find("a.close")
		.on("click", remove);

	setTimeout(remove, timeout || defaultTimeout);
}

function findOffset($el) {
	var offset = parseInt($el.css("top"), 10);

	$(".alert-notify").each((i, el) => {
		if (i === 0) return;
		offset += $(el).outerHeight() + 10;
	});

	return offset;
}

/*
 * @description Defines 4 functions that display 4 types of notifications.
 * Available notification types are: success, warning, error and info.
 * @param {string} msg - The message to display
 * @param {number=} timeout - Timeout in milliseconds. Defaults to 3s.
 */
["success", "warning", "error", "info"].forEach(type => {
	module.exports[type] = (msg, timeout) => displayNotification(type, msg, timeout);
});
