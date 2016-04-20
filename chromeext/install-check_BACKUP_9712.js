/**
The Sourcegraph application includes an inline installation CTA
(https://developer.chrome.com/webstore/inline_installation) which we
only want to show if the chrome extension is NOT installed.

The best [sic] way to tell Sourcegraph that the chrome extension is installed
is by indirectly adding a DOM element to the page which Sourcegraph
can detect in order to conditionally show the CTA.

See: https://groups.google.com/a/chromium.org/forum/#!topic/chromium-extensions/8ArcsWMBaM4
*/

function detectExistence() {
<<<<<<< 0dcab106d2c23b01077a0b0b69b34725dd6dac78
	console.log("we're in shte script");
	var el = document.createElement("div");
	el.id = "chrome-extension-installed";
	document.getElementById("chrome-extension-install-button").appendChild(el);
};
=======
	var el = document.createElement("div");
	el.id = "chrome-extension-installed";
	document.getElementById("chrome-extension-install-button").appendChild(el);
}
>>>>>>> fall back and trigger builds

document.addEventListener("DOMContentLoaded", detectExistence);
