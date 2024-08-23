# Changelog

This file tracks changes to this project. It follows the [Keep a Changelog format](https://keepachangelog.com/en/1.0.0/), and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.8.0] - 2024-01-30

This version removes all external dependencies of `go-sse`. All our bugs are belong to us! It also does some API and documentation cleanups.

### Removed

- `Client.DefautReconnectionTime`, `Client.MaxRetries` have been replaced with the new `Client.Backoff` configuration field. See the Added section for more info.
- `ErrReplayFailed` is removed from the public API.
- `ReplayProviderWithGC` and `Joe.ReplayGCInterval` are no more. The responsibility for garbage collection is assigned to the replay providers.

### Changed

- `Server.Logger` is now of a new type: the `Logger` interface. The dependency on x/exp/slog is removed. This opens up the possibility to adapt any existing logger to be usable with `Server`.
- The default backoff behavior has changed. The _previous_ defaults map to the new `Backoff` configuration as follows:
```go
sse.Backoff{
    InitialInterval:    5 * time.Second,  // currently 500ms
    Multiplier:         1.5,              // currently the same
    Jitter:             0.5,              // currently the same
    MaxInterval:        60 * time.Second, // currently unbounded
    MaxElapsedDuration: 15 * time.Minute, // currently unbounded
    MaxRetries:         -1,               // previously no retries by default, currently unbounded
}
```
- `Joe` now accepts new subscriptions even if replay providers panic (previously `ErrReplayFailed` would be returned).
- `Server.ServeHTTP` panics if a custom `OnSession` handler returns a `Subscription` with 0 topics

### Added

- The `Logger` interface, `LogLevel` type, and `LogLevel(Info|Warn|Error)` values.
- `Backoff` and `Client.Backoff` – the backoff strategy is now fully configurable. See the code documentation for info.
- `ValidReplayProvider.GCInterval`, to configure at which interval expired events should be cleaned up.

## [0.7.0] - 2023-11-19

This version overhauls connection retry and fixes the connection event dispatch order issue. Some internal changes to Joe were also made, which makes it faster and more resilient.

### Removed

- `ConnectionError.Temporary`
- `ConnectionError.Timeout`

### Changed

- Go's `Timeout` and `Temporary` interfaces are not used anymore – the client makes no assumptions and retries on every network or response read error. The only cases when `Connection.Connect` returns now are either when there are no more retries left (when the number is not infinite), or when the request context was cancelled.
- `*url.Error`s that occur on the HTTP request are now unwrapped and their cause is put inside a `ConnectionError`.
- `Connection.Connect` doesn't suppress any errors anymore: the request context errors are returned as is, all other errors are wrapped inside `ConnectionError`.
- On reconnection attempt, the response reset error is now wrapped inside `ConnectionError`. With this change, all errors other than the context errors are wrapped inside `ConnectionError`.
- Subscription callbacks are no longer called in individual goroutines. This caused messages to be received in an indereminate order. Make sure that your callbacks do not block for too long!

### Changed

- If a `ReplayProvider` method panics when called by `Joe`, instead of closing itself completely it just stops replaying, putting or GC-ing messages to upcoming clients. `Joe` continues to function as if no replay provider was given. A stack trace is printed to stderr when such a panic occurs.

## [0.6.0] - 2023-07-22

This version brings a number of refactors to the server-side tooling the library offers. Constructors and construction related types are removed, for ease of use and reduced API size, concerns regarding topics and expiry were separated from `Message`, logging of the `Server` is upgraded to structured logging and messages can be now published to multiple topics at once. Request upgrading has also been refactored to provide a more functional API, and the `Server` logic can now be customized without having to create a distinct handler.

### Removed

- `Message.ExpiresAt` is no more.
- `Message.Topic` is no more. See the changes to `Server`, `Provider` and `ReplayProvider` for handling topics – you can now publish a message to multiple topics at once.
- `Message.Writer` is no more. The API was redundant – one can achieve the same using `strings.Builder` and `Message.AppendData`. See the `MessageWriter` example for more.
- `NewValidReplayProvider` is no more.
- `NewFiniteReplayProvider` is no more.
- `NewJoe` is no more.
- `JoeConfig` is no more.
- `Server.Subscribe` is no more – it never made sense.
- `Server.Provider` is no more.
- `NewServer`, `ServerOption` and friends are no more.
- The `Logger` interface and the capability of the `Server` to use types that implement `Logger` as logging systems is removed.
- `SubscriptionCallback` is no more (see the change to the `Subscription` type in the "Changed" section).

### Added

- Because the `ValidReplayProvider` constructor was removed, the fields `ValidReplayProvider.{TTL,AutoIDs}` were added for configuration.
- Because the `FiniteReplayProvider` constructor was removed, the fields `FiniteReplayProvider.{Count,AutoIDs}` were added for configuration.
- Because the `Joe` constructor was removed, the fields `Joe.{ReplayProvider,ReplayGCInterval}` were added for configuration.
- Because the `Server` constructor was removed, the field `Server.Provider` was added for configuration.
- New `MessageWriter` interface; used by providers to send messages and implemented by `Session` (previously named `Request`).
- New `ResponseWriter` interface, which is a `http.ResponseWriter` augmented with a `Flush` method.
- `ValidReplayProvider` has a new field `Now` which allows providing a custom current time getter, like `time.Now`, to the provider. Enables deterministic testing of dependents on `ValidReplayProvider`.
- New `Server.OnSession` field, which enables customization of `Server`'s response and subscriptions.
- New `Server.Logger` field, which enables structured logging with logger retrieved from the request and customizable config of logged information.

### Changed

- `ReplayProvider.Put` takes a simple `*Message` and returns a `*Message`, instead of changing the `*Message` to which the `**Message` parameter points.
  It also takes a slice of topics, given that the `Message` doesn't hold the topic itself anymore. If the Message cannot be put, the method must now panic – see documentation for info.
- Because `Message.ExpiresAt` is removed, the `ValidReplayProvider` sets the expiry itself.
- `Server.Publish` now takes a list of topics.
- `Provider.Publish` now takes a non-empty slice of topics.
- `ReplayProvider.Put` now takes a non-empty slice of topics.
- `Provider.Stop` is now `Provider.Shutdown` and takes now a `context.Context` as a parameter.
- `Server.Shutdown` takes now a `context.Context` as a parameter.
- `Request` is now named `Session` and exposes the HTTP request, response writer, and the last event ID of the request.
- A new method `Flush` is added to `Session`; messages are no longer flushed by default, which allows providers, replay providers to batch send messages.
- `Upgrade` now takes an `*http.Request` as its second parameter.
- `Subscription` now has a `Client` field of type `MessageWriter` instead of a `Callback`.
- Given the `Subscription` change, `Provider.Subscribe` and `ReplayProvider.Replay` now report message sending errors.


## [0.5.2] - 2023-07-12

### Added

- The new `Message.Writer` – write to the `Message` as if it is an `io.Writer`.

### Fixed

- `Message.UnmarshalText` now strips the leading Unicode BOM, if it exists, as per the specification.
- When parsing events client-side, BOM removal was attempted on each event input. Now the BOM is correctly removed only when parsing is started.

## [0.5.1] - 2023-07-12

### Fixed

- `Message.WriteTo` now writes nothing if `Message` is empty.
- `Message.WriteTo` does not attempt to write the `retry` field if `Message.Retry` is not at least 1ms.
- `NewType` error message is updated to say "event type", not "event name".

## [0.5.0] - 2023-07-11

This version comes with a series of internal refactorings that improve code readability and performance. It also replaces usage of `[]byte` for event data with `string` – SSE is a UTF-8 encoded text-based protocol, so raw bytes never made sense. This migration improves code safety (less `unsafe` usage and less worry about ownership) and reduces the memory footprint of some objects.

Creating events on the server is also revised – fields that required getters and setters, apart from `data` and comments, are now simple public fields on the `sse.Message` struct.

Across the codebase, to refer to the value of the `event` field the name "event type" is used, which is the nomenclature used in the SSE specification.

Documentation and examples were also fixed and improved.

### Added

- New `sse.EventName` type, which holds valid values for the `event` field, together with constructors (`sse.Name` and `sse.NewName`).

### Removed

- `sse.Message`: `AppendText` was removed, as part of the migration from byte slices to strings. SSE is a UTF-8 encoded text-based protocol – raw bytes never made sense.

### Changed

- Minimum supported Go version was bumped from 1.16 to 1.19. From now on, the latest two major Go versions will be supported.
- `sse.Message`: `AppendData` takes `string`s instead of `[]byte`.
- `sse.Message`: `Comment` is now named `AppendComment`, for consistency with `AppendData`.
- `sse.Message`: The message's expiration is not reset anymore by `UnmarshalText`.
- `sse.Message`: `UnmarshalText` now unmarshals comments aswell.
- `sse.Message`: `WriteTo` (and `MarshalText` and `String` as a result) replaces all newline sequences in data with LF.
- `sse.Message`: The `Expiry` getter and `SetExpiresAt`, `SetTTL` setters are replaced by the public field `ExpiresAt`.
- `sse.Message`: Event ID getter and setter are replaced by the public `ID` field.
- `sse.Message`: Event type (previously named `Name`) getter and setter are replaced by the public `Type` field.
- `sse.Message`: The `retry` field value is now a public field on the struct. As a byproduct, `WriteTo` will now make 1 allocation when writing events with the `retry` field set. 
- `sse.NewEventID` is now `sse.NewID`, and `sse.MustEventID` is `sse.ID`.
- `sse.Event`: The `Data` field is now of type `string`, not `[]byte`.
- `sse.Event`: The `Name` field is now named `Type`.

### Fixed

- `sse.Message`: `Clone` now copies the topic of the message to the new value.
- `sse.Message`: ID fields that contain NUL characters are now ignored, as required by the spec, in `UnmarshalText`.

## [0.4.3] - 2023-07-08

### Fixed

- Messages longer than 4096 bytes are no longer being dropped ([#2], thanks [@aldld])
- Event parsing no longer panics on empty field with colon after name, see [test case](https://github.com/tmaxmax/go-sse/blob/4938f99db3bf7a8f057cb3e21ca88df57db3c0e0/internal/parser/field_parser_test.go#L37-L45) for example ([#5])

## [0.4.2] - 2021-10-17

### Added

- Get the event name of a Message

## [0.4.1] - 2021-10-15

### Added

- Set a custom logger for Server

## [0.4.0] - 2021-10-15

### Changed

- Server does not set any other headers besides `Content-Type`.
- UpgradedRequest does not return a SendError anymore when Write errors.
- Providers don't handle callback errors anymore. Callbacks return a flag that indicates whether the provider should keep calling it for new messages instead.

### Fixed

- Client's default response validator now ignores `Content-Type` parameters when checking if the response's content type is `text/event-stream`.
- Various optimizations

## [0.3.0] - 2021-09-18

### Added

- ReplayProviderWithGC interface, which must be satisfied by replay providers that must be cleaned up periodically.

### Changed

- Subscriptions now take a callback function instead of a channel.
- Server response headers are now sent on the first Send call, not when Upgrade is called.
- Providers are not required to add the default topic anymore. Callers of Subscribe should ensure at least a topic is specified.
- Providers' Subscribe method now blocks until the subscriber is removed.
- Server's Subscribe method automatically adds the default topic if no topic is specified.
- ReplayProvider does not require for GC to be implemented.
- Client connections take callback functions instead of channels as event listeners.
- Client connections' Unsubscribe methods are replaced by functions returned by their Subscribe counterparts.

### Fixed

- Fix replay providers not replaying the oldest message if the ID provided is of the one before that one.
- Fix replay providers hanging the caller's goroutine when a write error occurs using the default ServeHTTP implementation.
- Fix providers hanging when a write error occurs using the default ServeHTTP implementation.

## [0.2.0] - 2021-09-13

### Added

- Text/JSON marshalers and unmarshalers, and SQL scanners and valuers for the EventID type (previously event.ID).
- Check for http.NoBody before resetting the request body on client reconnect.

### Changed

- Package structure. The module is now refactored into a single package with an idiomatic name. This has resulted in various name changes:
  - `client.Error` - `sse.ConnectionError`
  - `event.Event` - `sse.Message` (previous `server.Message` is removed, see next change)
  - `event.ID` - `sse.EventID`
  - `event.NewID` - `sse.NewEventID`
  - `event.MustID` - `sse.MustEventID`
  - `server.Connection` - `sse.UpgradedRequest`
  - `server.NewConnection` - `sse.Upgrade`
  - `server.ErrUnsupported` - `sse.ErrUpgradeUnsupported`
  - `server.New` - `sse.NewServer`.
- `event.Event` is merged with `server.Message`, becoming `sse.Message`. This affects the `sse.Server.Publish` function, which doesn't take a `topic` parameter anymore.
- The server's constructor doesn't take an `Provider` as a parameter. It instead takes multiple optional `ServerOptions`. The `WithProvider` option is now used to pass custom providers to the server.
- The `ReplayProvider` interface's `Put` method now takes a `**Message` instead of a `*Message`. This change also affects the replay providers in this package: `ValidReplayProvider` and `FiniteReplayProvider`.
- The `Provider` interface's `Publish` method now takes a `*Message` instead of a `Message`. This change also affects `Joe`, the provider in this package.
- The `UpgradedRequest`'s `Send` now method takes a `*Message` as parameter.

## [0.1.0] - 2021-09-11 First release

[@aldld]: https://github.com/aldld

[#5]: https://github.com/tmaxmax/go-sse/pull/5
[#2]: https://github.com/tmaxmax/go-sse/pull/2

[0.6.0]: https://github.com/tmaxmax/go-sse/releases/tag/v0.6.0
[0.5.2]: https://github.com/tmaxmax/go-sse/releases/tag/v0.5.2
[0.5.1]: https://github.com/tmaxmax/go-sse/releases/tag/v0.5.1
[0.5.0]: https://github.com/tmaxmax/go-sse/releases/tag/v0.5.0
[0.4.3]: https://github.com/tmaxmax/go-sse/releases/tag/v0.4.3
[0.4.2]: https://github.com/tmaxmax/go-sse/releases/tag/v0.4.2
[0.4.1]: https://github.com/tmaxmax/go-sse/releases/tag/v0.4.1
[0.4.0]: https://github.com/tmaxmax/go-sse/releases/tag/v0.4.0
[0.3.0]: https://github.com/tmaxmax/go-sse/releases/tag/v0.3.0
[0.2.0]: https://github.com/tmaxmax/go-sse/releases/tag/v0.2.0
[0.1.0]: https://github.com/tmaxmax/go-sse/releases/tag/v0.1.0
