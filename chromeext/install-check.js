/**
The Sourcegraph application includes an inline installation CTA
(https://developer.chrome.com/webstore/inline_installation) which we
only want to show if the chrome extension is NOT installed.

The best [sic] way to tell Sourcegraph that the chrome extension is installed
is by indirectly adding a DOM element to the page which Sourcegraph
can detect in order to conditionally show the CTA.

See: https://groups.google.com/a/chromium.org/forum/#!topic/chromium-extensions/8ArcsWMBaM4
*/

function addScript() {
	var script = document.createElement("script");
	script.id = "chrome-extension-installed";
	document.body.appendChild(script);
}

document.addEventListener("DOMContentLoaded", addScript);
