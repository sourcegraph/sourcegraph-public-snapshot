declare var AJS:
    | {
          /**
           * The AJS.contextPath() function returns the "path" to the application,
           * which is needed when creating absolute urls within the application.
           */
          contextPath(): string
      }
    | undefined
