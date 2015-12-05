package notifications

import "src.sourcegraph.com/apps/notifications/notifications"

// Service is the external API of Notifications Center app.
//
// Its value is set during app startup, and should only be accessed after all
// apps have finished initializing.
//
// Its value is nil if Notification Center app is not started.
//
// TODO: Try this out initially, see if this can/should be made better.
var Service notifications.ExternalService = nil
