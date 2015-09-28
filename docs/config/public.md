+++
title = "Public server"
+++

By default, only specific users (chosen by the administrator) may
access a Sourcegraph server. To run a publicly accessible Sourcegraph
server (for open-source projects, for example), you can use the
following configuration settings.

```
[app.Authentication]
; Allow anonymous (non-logged-in) users to perform read operations
; (e.g., browsing code).
AllowAnonReaders = true

; Allow all users, not just explicitly granted users, to log in.
AllowAllLogins = true

; Only allow git push, repo updates, changeset creation, etc.,
; by users with the 'admin' permission grant.
RestrictWriteAccess = true
```
