var script = document.createElement('script');
script.type = 'text/javascript';
script.defer = true;
script.src = window.SOURCEGRAPH_URL + '/.assets/extension/scripts/phabricator.bundle.js';
document.getElementsByTagName('head')[0].appendChild(script);

var head = document.head || document.getElementsByTagName('head')[0];
var styleLink = document.createElement('link');
styleLink.rel = 'stylesheet';
styleLink.href = window.SOURCEGRAPH_URL + '/.assets/extension/css/style.bundle.css';
head.appendChild(styleLink);
