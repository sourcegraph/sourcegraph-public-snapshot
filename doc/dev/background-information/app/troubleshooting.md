# Troubleshooting App not loading

If you have issues with the app not loading (opening the window but the window stays blank), try the following:

1. Check the error console by clicking on the tray icon and clicking "Troubleshoot"
2. If the error is related to PostgreSQL, delete the following folders:
    - `~/Library/Application Support/com.sourcegraph.cody`
    - `~/.sourcegraph-psql`

If you are still having problems, please contact #ask-app in Slack (for Sourcegraph teammates) or #app in Discord (for Sourcegraph community members).
