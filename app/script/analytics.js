// This file is loaded inline in template/layout.html

var $ = require("jquery");

/**
 * @description Maps document header presence to analytics methods. If the key correctly
 * describes a header in document.head.dataset, the corresponding method will be invoked.
 * @type {Object}
 */
var analyticsHeaders = {
	heapAnalyticsId: setupHeapAnalytics,
	googleAnalyticsTrackingId: setupGoogleAnalytics,
	publicRavenDsn: setupSentry,
};

$(function() {
	Object.keys(analyticsHeaders).forEach(function(header) {
		if (document.head.dataset.hasOwnProperty(header)) analyticsHeaders[header]();
	});
});

/**
 * @description Sets up HEAP Analytics.
 * @returns {void}
 */
function setupHeapAnalytics() {
	if (!document.head.dataset.heapAnalyticsId) {
		return;
	}

	window.heap=window.heap||[],heap.load=function(t,e){window.heap.appid=t,window.heap.config=e;var a=document.createElement("script");a.type="text/javascript",a.async=!0,a.src="https://cdn.heapanalytics.com/js/heap-"+t+".js";var n=document.getElementsByTagName("script")[0];n.parentNode.insertBefore(a,n);for(var o=function(t){return function(){heap.push([t].concat(Array.prototype.slice.call(arguments,0)))}},p=["clearEventProperties","identify","setEventProperties","track","unsetEventProperty"],c=0;c<p.length;c++)heap[p[c]]=o(p[c])}; // eslint-disable-line

	window.heap.load(document.head.dataset.heapAnalyticsId);

	if (document.head.dataset.currentUserLogin) {
		window.heap.identify({handle: document.head.dataset.currentUserLogin});
	}

	document.addEventListener("DOMContentLoaded", function() {
		//
		// Custom page view tracking by route name
		//
		function trackPageView() {
			$.get("/_/route/" + window.location.pathname, null, function(routeName) {
				var properties = {path: window.location.pathname, query: window.location.search};

				var statusCodeElem = $("#sg-status-code");
				if (statusCodeElem.size() >= 1) {
					properties.statusCode = statusCodeElem.html();
				}

				if (routeName === "search" || routeName.indexOf("search.") === 0) {
					if ($(".no-results-found").size() >= 1) {
						properties.noSearchResults = true;
					}
				}

				window.heap.track("Page(route): "+routeName, properties);
			});
		}

		trackPageView();

		["pushState", "popState", "replaceState"].forEach(function(method) {
			window.addEventListener("sg:"+method, function(e) {
				trackPageView();
			});
		});
	});
}

/**
 * @description Enables google analytics tracking.
 * @returns {void}
 */
function setupGoogleAnalytics() {
	if (!document.head.dataset.googleAnalyticsTrackingId) {
		return;
	}

	(function(i,s,o,g,r,a,m){i["GoogleAnalyticsObject"]=r;i[r]=i[r]||function(){(i[r].q=i[r].q||[]).push(arguments)},i[r].l=1*new Date();a=s.createElement(o),m=s.getElementsByTagName(o)[0];a.async=1;a.src=g;m.parentNode.insertBefore(a,m)})(window,document,"script","https://www.google-analytics.com/analytics.js","ga"); // eslint-disable-line

	window.ga("create", document.head.dataset.googleAnalyticsTrackingId, "auto");
	window.ga("require", "linkid", "https://www.google-analytics.com/plugins/ua/linkid.js");
	window.ga("set", "dimension1", Boolean(document.head.dataset.currentUserLogin));
	window.ga("set", "dimension4", "web");
	window.ga("send", "pageview");

	["pushState", "popState", "replaceState"].forEach(function(method) {
		window.addEventListener("sg:"+method, function(e) {
			window.ga("send", "pageview", window.location.pathname);
		});
	});
}

/**
 * @description Sets up AppSentry error reporting.
 * @returns {void}
 */
function setupSentry() {
	document.addEventListener("DOMContentLoaded", function() {
		require("raven-js/dist/raven.js");
		require("raven-js/plugins/native.js");
		require("raven-js/plugins/console.js");

		// When loading plugins, load the libs they modify before loading
		// the plugins, or else the plugins can't take effect.
		require("jquery");
		require("raven-js/plugins/jquery.js");
		require("backbone");
		require("raven-js/plugins/backbone.js");

		// Ignore rules (from https://gist.github.com/impressiver/5092952).
		var opt = {
			// Will cause a deprecation warning, but the demise of `ignoreErrors` is still under discussion.
			// See: https://github.com/getsentry/raven-js/issues/73
			ignoreErrors: [
				// Random plugins/extensions
				"top.GLOBALS",
				// See: http://blog.errorception.com/2012/03/tale-of-unfindable-js-error.html
				"originalCreateNotification",
				"canvas.contentDocument",
				"MyApp_RemoveAllHighlights",
				"http://tt.epicplay.com",
				"Can\'t find variable: ZiteReader",
				"jigsaw is not defined",
				"ComboSearch is not defined",
				"http://loading.retry.widdit.com/",
				"atomicFindClose",
				// Facebook borked
				"fb_xd_fragment",
				// ISP "optimizing" proxy - `Cache-Control: no-transform` seems to reduce this. (thanks @acdha)
				// See http://stackoverflow.com/questions/4113268/how-to-stop-javascript-injection-from-vodafone-proxy
				"bmi_SafeAddOnload",
				"EBCallBackMessageReceived",
				// See http://toolbar.conduit.com/Developer/HtmlAndGadget/Methods/JSInjection.aspx
				"conduitPage",
				// Generic error code from errors outside the security sandbox
				// You can delete this if using raven.js > 1.0, which ignores these automatically.
				"Script error.",
			],
			ignoreUrls: [
				// Facebook flakiness
				/graph\.facebook\.com/i,
				// Facebook blocked
				/connect\.facebook\.net\/en_US\/all\.js/i,
				// Woopra flakiness
				/eatdifferent\.com\.woopra-ns\.com/i,
				/static\.woopra\.com\/js\/woopra\.js/i,
				// Chrome extensions
				/extensions\//i,
				/^chrome:\/\//i,
				// Other plugins
				/127\.0\.0\.1:4001\/isrunning/i,  // Cacaoweb
				/webappstoolbarba\.texthelp\.com\//i,
				/metrics\.itunes\.apple\.com\.edgesuite\.net\//i,
			],
		};
		opt.tags = {};

		if (document.head.dataset.deployedGitCommitId) {
			opt.tags["Deployed commit"] = document.head.dataset.deployedGitCommitId;
		}
		opt.tags["Authed"] = document.head.dataset.currentUserLogin ? "yes" : "no";
		if (document.head.dataset.currentUserUid && document.head.dataset.currentUserLogin) {
			opt.tags["Authed UID"] = document.head.dataset.currentUserUid;
			opt.tags["Authed user"] = document.head.dataset.currentUserLogin;
		}
		if (document.head.dataset.appdashCurrentSpanIdTrace) {
			opt.tags["Appdash trace"] = document.head.dataset.appdashCurrentSpanIdTrace;
		}
		if (document.head.dataset.appdashCurrentSpanIdSpan) {
			opt.tags["Appdash span"] = document.head.dataset.appdashCurrentSpanIdSpan;
		}
		if (document.head.dataset.appdashCurrentSpanIdParent) {
			opt.tags["Appdash parent"] = document.head.dataset.appdashCurrentSpanIdParent;
		}

		window.Raven.config(document.head.dataset.publicRavenDsn, opt).install();
	});
}
