# repo-updater

Repo-updater (which may get renamed since it does more than that) tracks the state of repos, and is responsible for automatically scheduling updates ("git fetch" runs) using gitserver. Other apps which desire updates or fetches should be telling repo-updater, rather than using gitserver directly, so repo-updater can take their changes into account. Only one replica should be running.
