# How to troubleshoot a Sourcegraph extension

This guide gives specific instructions for troubleshooting [extensions](https://docs.sourcegraph.com/extensions) developed by Sourcegraph.

## Prerequisites

* This document assumes that you already have the [Sourcegraph browser extension](https://docs.sourcegraph.com/integration/browser_extension) installed

## FAQs

#### How do I know if a Sourcegraph extension is running?

1. Right click on the Sourcegraph website and click Inspect (Chrome/Firefox) / Inspect Element (Safari) to open `Developer Tools`
2. You should see a console message `Activating Sourcegraph extension:` follows by the names of all the running extensions in the `Console` tab
3. If you don't see the expected extension running, please go to the User Menu on your Sourcegraph instance and click on Extensions to make sure the extension is enabled

#### A Sourcegraph extension is not working, what should I do?

1. First of all, please make sure the extension in question is running by following the steps from above
2. Look for error messages in your browser's `Developer Console`
3. Look for error messages in your browser's  `Network panel`

#### sonarqube: Error fetching Sonarqube data: Error: Forbidden
1. Look for error messages in your browser's  `Network panel`
2. If there is an error message indicates that the cors-anywhere request was being denied and you are using `"sonarqube.corsAnywhereUrl": "https://cors-anywhere.herokuapp.com"` in your configuration, please visit [https://cors-anywhere.herokuapp.com/corsdemo](https://cors-anywhere.herokuapp.com/corsdemo) to opt-in for temporary access by clicking on the `Request temporary access to the demo server` button
3. Alternatively you may remove this configuration option to use the default Sourcegraph's CORS proxy

#### Definition from the Code-Intel language extensions are slow to load 

#### Why is the extension icon on my sidebar is shown as inactive / greyed out?

#### Git blame is not working

#### Git blame is slow to load

#### Can I export search results?
1. You can export search results by enabling the [Sourcegraph search results CSV export extension](https://sourcegraph.com/extensions/sourcegraph/search-export)
2. Once it is enabled, you will find an `Export to CSV` button in the Search-Results page


See the [Sourcegraph browser extension docs](https://docs.sourcegraph.com/integration/browser_extension#troubleshooting) for more troubleshooting tips.
