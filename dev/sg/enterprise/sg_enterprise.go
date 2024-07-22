// Package enterprise exports 'sg enterprise' commands for interacting with the
// Sourcegraph Enterprise Portal service.
package enterprise

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"connectrpc.com/connect"
	"github.com/urfave/cli/v2"
	"golang.org/x/exp/maps"
	"golang.org/x/oauth2"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/fieldmaskpb"

	sams "github.com/sourcegraph/sourcegraph-accounts-sdk-go"
	"github.com/sourcegraph/sourcegraph-accounts-sdk-go/scopes"
	subscriptionsv1 "github.com/sourcegraph/sourcegraph/lib/enterpriseportal/subscriptions/v1"
	subscriptionsv1connect "github.com/sourcegraph/sourcegraph/lib/enterpriseportal/subscriptions/v1/v1connect"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/pointers"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/category"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/dev/sg/sams/samsflags"
)

const (
	enterprisePortalProdURL = "https://enterprise-portal.sourcegraph.com"
	enterprisePortalDevURL  = "https://enterprise-portal.sgdev.org"
)

var (
	scopeWriteSubscriptions = scopes.ToScope(scopes.ServiceEnterprisePortal, "subscription", scopes.ActionWrite)
	scopeReadSubscriptions  = scopes.ToScope(scopes.ServiceEnterprisePortal, "subscription", scopes.ActionRead)

	scopeWriteSubscriptionsPermissions = scopes.ToScope(scopes.ServiceEnterprisePortal, "permission.subscription", scopes.ActionWrite)
)

func clientFlags() []cli.Flag {
	return append(samsflags.ClientCredentials(), &cli.StringFlag{
		Name:    "enterprise-portal-server",
		Aliases: []string{"server"},
		Usage:   "The URL of the Enterprise Portal server to use (defaults to the appropriate one for SG_SAMS_SERVER_URL)",
	})
}

func newSubscriptionsClient(c *cli.Context, ss ...scopes.Scope) (subscriptionsv1connect.SubscriptionsServiceClient, error) {
	ctx := c.Context
	samsServer := c.String("sams-server")
	enterprisePortal := enterprisePortalDevURL
	if epServer := c.String("enterprise-portal-server"); epServer != "" {
		enterprisePortal = epServer
	} else if samsServer == samsflags.SAMSProdURL {
		enterprisePortal = enterprisePortalProdURL
	}

	std.Out.WriteSuggestionf("Using %q and %q",
		enterprisePortal, samsServer)

	samsCfg, err := samsflags.NewClientCredentialsFromFlags(c, ss)
	if err != nil {
		return nil, err
	}
	return subscriptionsv1connect.NewSubscriptionsServiceClient(
		oauth2.NewClient(ctx, samsCfg.TokenSource(ctx)),
		enterprisePortal), err
}

// resolveUserReference converts a 'user reference' provided as an argument or
// flag, into a SAMS account ID. The 'user reference' can either be a SAMS user
// ID, or an email address (determined by the presence of '@' which is also an
// illegal character in a SAMS user ID).
//
// Required scope: profile
func resolveUserReference(ctx context.Context, users *sams.UsersServiceV1, userReference string) (samsAccountID string, _ error) {
	if strings.Contains(userReference, "@") {
		user, err := users.GetUserByEmail(ctx, userReference)
		if err != nil {
			return "", errors.Wrapf(err, "get user by email %q", userReference)
		}
		samsAccountID = user.Id
	} else {
		user, err := users.GetUserByID(ctx, userReference) // check if it's a valid user ID
		if err != nil {
			return "", errors.Wrapf(err, "get user by ID %q", userReference)
		}
		samsAccountID = user.GetId()
	}

	if samsAccountID != userReference {
		std.Out.WriteSuggestionf("Resolved user %q to %q", userReference, samsAccountID)
	}

	return samsAccountID, nil
}

// Command is the 'sg sams' toolchain for the Sourcegraph Accounts Management System (SAMS).
var Command = &cli.Command{
	Name:     "enterprise",
	Category: category.Company,
	Usage:    "[EXPERIMENTAL] Manage Sourcegraph Enterprise subscriptions in Enterprise Portal",
	Description: `[EXPERIMENTAL] See https://www.notion.so/sourcegraph/Sourcegraph-Enterprise-Portal-EP-c6311818978547beb981bd2a593c6acf?pvs=4

Please reach out to #discuss-core-services for assistance if you have any questions!`,
	Subcommands: []*cli.Command{{
		Name:  "subscription",
		Usage: "Manage Sourcegraph Enterprise subscriptions in Enterprise Portal",
		Subcommands: []*cli.Command{{
			Name:      "list",
			Usage:     "List subscriptions",
			ArgsUsage: "[subscription IDs...]",
			Flags: append(clientFlags(),
				&cli.StringFlag{
					Name:  "member.cody-analytics-viewer",
					Usage: "Member with Cody Analytics viewer permission to filter for (email or SAMS user ID)",
				}),
			Action: func(c *cli.Context) error {
				client, err := newSubscriptionsClient(c, scopeReadSubscriptions)
				if err != nil {
					return err
				}
				req := &subscriptionsv1.ListEnterpriseSubscriptionsRequest{
					Filters: []*subscriptionsv1.ListEnterpriseSubscriptionsFilter{{
						Filter: &subscriptionsv1.ListEnterpriseSubscriptionsFilter_IsArchived{
							IsArchived: false,
						},
					}},
				}
				for _, subscription := range c.Args().Slice() {
					if !strings.HasPrefix(subscription, subscriptionsv1.EnterpriseSubscriptionIDPrefix) {
						return errors.Newf("invalid subscription ID %q, expected prefix %q",
							subscription, subscriptionsv1.EnterpriseSubscriptionIDPrefix)
					}
					req.Filters = append(req.Filters, &subscriptionsv1.ListEnterpriseSubscriptionsFilter{
						Filter: &subscriptionsv1.ListEnterpriseSubscriptionsFilter_SubscriptionId{
							SubscriptionId: subscription,
						},
					})
				}
				if member := c.String("member.cody-analytics-viewer"); member != "" {
					samsClient, err := samsflags.NewClientFromFlags(c, scopes.Scopes{scopes.Profile})
					if err != nil {
						return errors.Wrap(err, "get SAMS client")
					}
					samsUserID, err := resolveUserReference(c.Context, samsClient.Users(), member)
					if err != nil {
						return errors.Wrap(err, "resolve SAMS user ID")
					}
					req.Filters = append(req.Filters, &subscriptionsv1.ListEnterpriseSubscriptionsFilter{
						Filter: &subscriptionsv1.ListEnterpriseSubscriptionsFilter_Permission{
							Permission: &subscriptionsv1.Permission{
								Type:          subscriptionsv1.PermissionType_PERMISSION_TYPE_SUBSCRIPTION_CODY_ANALYTICS,
								Relation:      subscriptionsv1.PermissionRelation_PERMISSION_RELATION_VIEW,
								SamsAccountId: samsUserID,
							},
						},
					})
				}

				resp, err := client.ListEnterpriseSubscriptions(c.Context, connect.NewRequest(req))
				if err != nil {
					return errors.Wrap(err, "list subscriptions")
				}

				for _, s := range resp.Msg.GetSubscriptions() {
					data, err := protojson.MarshalOptions{
						Multiline: true,
					}.Marshal(s)
					if err != nil {
						std.Out.WriteWarningf("Failed to marshal subscription %q: %s",
							s.GetId(), err.Error())
						continue
					}
					_ = std.Out.WriteCode("json", string(data))
				}
				std.Out.WriteSuccessf("Found %d subscriptions", len(resp.Msg.GetSubscriptions()))
				return nil
			},
		}, {
			Name:  "license",
			Usage: "Manage Enterprise subscription licenses",
			Subcommands: []*cli.Command{{
				Name:      "list",
				Usage:     "List licenses for all or specified subscriptions",
				ArgsUsage: "[subscription IDs...]",
				Flags:     clientFlags(),
				Action: func(c *cli.Context) error {
					client, err := newSubscriptionsClient(c, scopeReadSubscriptions)
					if err != nil {
						return err
					}
					req := &subscriptionsv1.ListEnterpriseSubscriptionLicensesRequest{
						Filters: []*subscriptionsv1.ListEnterpriseSubscriptionLicensesFilter{{
							Filter: &subscriptionsv1.ListEnterpriseSubscriptionLicensesFilter_IsRevoked{
								IsRevoked: false,
							},
						}},
					}
					for _, subscription := range c.Args().Slice() {
						if !strings.HasPrefix(subscription, subscriptionsv1.EnterpriseSubscriptionIDPrefix) {
							continue
						}
						req.Filters = append(req.Filters, &subscriptionsv1.ListEnterpriseSubscriptionLicensesFilter{
							Filter: &subscriptionsv1.ListEnterpriseSubscriptionLicensesFilter_SubscriptionId{
								SubscriptionId: subscription,
							},
						})
					}
					resp, err := client.ListEnterpriseSubscriptionLicenses(c.Context, connect.NewRequest(req))
					if err != nil {
						return errors.Wrap(err, "list subscriptions")
					}
					for _, s := range resp.Msg.GetLicenses() {
						if k := s.GetKey(); k != nil {
							k.LicenseKey = "<redacted>"
						}
						data, err := protojson.MarshalOptions{
							Multiline: true,
						}.Marshal(s)
						if err != nil {
							std.Out.WriteWarningf("Failed to marshal license %q: %s",
								s.GetId(), err.Error())
							continue
						}
						_ = std.Out.WriteCode("json", string(data))
					}
					std.Out.WriteSuccessf("Found %d licenses", len(resp.Msg.GetLicenses()))
					return nil
				},
			}},
		}, {
			Name:      "set-instance-domain",
			Usage:     "Assign an instance domain to a subscription",
			ArgsUsage: "<subscription ID> <instance domain>",
			Flags: append(clientFlags(),
				&cli.BoolFlag{
					Name:  "auto-approve",
					Usage: "Skip confirmation prompts",
				}),
			Action: func(c *cli.Context) error {
				client, err := newSubscriptionsClient(c, scopeWriteSubscriptions)
				if err != nil {
					return err
				}
				s := &subscriptionsv1.EnterpriseSubscription{
					Id:             c.Args().Get(0),
					InstanceDomain: c.Args().Get(1),
				}
				if s.Id == "" {
					return errors.New("subscription ID required")
				}
				if !strings.HasPrefix(s.Id, subscriptionsv1.EnterpriseSubscriptionIDPrefix) {
					return errors.Newf("subscription ID must start with %q", subscriptionsv1.EnterpriseSubscriptionIDPrefix)
				}
				if s.InstanceDomain == "" && !c.Bool("auto-approve") {
					var res string
					ok, err := std.PromptAndScan(std.Out, "No instance domain provided; the assigned domain will be removed, are you sure? (y/N) ", &res)
					if err != nil {
						return err
					} else if !ok {
						return errors.New("response is required")
					}
					if res != "y" {
						return errors.New("aborting")
					}
				}
				s.InstanceDomain, err = subscriptionsv1.NormalizeInstanceDomain(s.InstanceDomain)
				if err != nil {
					return errors.Wrap(err, "normalize instance domain")
				}

				std.Out.Writef("Assigning domain %q to subscription %q\n",
					s.InstanceDomain, s.Id)
				resp, err := client.UpdateEnterpriseSubscription(c.Context, connect.NewRequest(&subscriptionsv1.UpdateEnterpriseSubscriptionRequest{
					Subscription: s,
					UpdateMask: &fieldmaskpb.FieldMask{
						Paths: []string{
							"instance_domain",
						},
					},
				}))
				if err != nil {
					return errors.Wrap(err, "update enterprise subscription")
				}

				updatedSub := resp.Msg.GetSubscription()
				std.Out.WriteSuccessf("Updated subscription %q with instance domain %q\n",
					pointers.Deref(pointers.NilIfZero(updatedSub.DisplayName), updatedSub.GetId()),
					updatedSub.GetInstanceDomain())
				return nil
			},
		}, {
			Name:      "set-name",
			Usage:     "Assign a display name name to a subscription",
			ArgsUsage: "<subscription ID> <display name>",
			Flags:     clientFlags(),
			Action: func(c *cli.Context) error {
				client, err := newSubscriptionsClient(c, scopeWriteSubscriptions)
				if err != nil {
					return err
				}
				s := &subscriptionsv1.EnterpriseSubscription{
					Id:          c.Args().Get(0),
					DisplayName: c.Args().Get(1),
				}
				if s.Id == "" {
					return errors.New("subscription ID required")
				}
				if !strings.HasPrefix(s.Id, subscriptionsv1.EnterpriseSubscriptionIDPrefix) {
					return errors.Newf("subscription ID must start with %q", subscriptionsv1.EnterpriseSubscriptionIDPrefix)
				}
				if s.DisplayName == "" {
					return errors.New("display name required")
				}

				std.Out.Writef("Assigning display name %q to subscription %q\n",
					s.DisplayName, s.Id)
				resp, err := client.UpdateEnterpriseSubscription(c.Context, connect.NewRequest(&subscriptionsv1.UpdateEnterpriseSubscriptionRequest{
					Subscription: s,
					UpdateMask: &fieldmaskpb.FieldMask{
						Paths: []string{
							"display_name",
						},
					},
				}))
				if err != nil {
					return errors.Wrap(err, "update enterprise subscription")
				}

				updatedSub := resp.Msg.GetSubscription()
				std.Out.WriteSuccessf("Updated subscription %q with display name %q\n",
					updatedSub.GetId(),
					updatedSub.GetDisplayName())
				return nil
			},
		}, {
			Name:        "update-membership",
			Usage:       "Update or assign membership to a subscription for one or more SAMS users",
			Description: "Only one of --subscription-id or --subscription-instance-domain needs to be specified.",
			ArgsUsage:   "<SAMS account email or ID...>",
			Flags: append(clientFlags(),
				&cli.StringFlag{
					Name:  "subscription-id",
					Usage: "ID of the subscription to assign membership to",
				},
				&cli.StringFlag{
					Name:  "subscription-instance-domain",
					Usage: "Assigned instance domain of the subscription to assign membership to",
				},
				&cli.StringSliceFlag{
					Name: "role",
					Usage: fmt.Sprintf("Roles to assign to the member - any of: [%s]",
						strings.Join(func() []string {
							values := maps.Clone(subscriptionsv1.Role_value)
							delete(values, "ROLE_UNSPECIFIED")
							keys := maps.Keys(values)
							slices.Sort(keys)
							return keys
						}(), ", ")),
				},
				&cli.BoolFlag{
					Name:  "auto-approve",
					Usage: "Skip confirmation prompts",
				}),
			Action: func(c *cli.Context) error {
				ctx := context.Background()
				client, err := newSubscriptionsClient(c,
					scopeWriteSubscriptions,
					scopeWriteSubscriptionsPermissions)
				if err != nil {
					return err
				}

				var (
					members = c.Args().Slice() // can be email or user ID
					roles   = c.StringSlice("role")

					subscriptionID             = c.String("subscription-id")
					subscriptionInstanceDomain = c.String("subscription-instance-domain")
				)
				if len(members) == 0 {
					return errors.New("at least one SAMS account email or ID is required")
				}
				if subscriptionID == "" && subscriptionInstanceDomain == "" {
					return errors.New("-subscription-id or -subscription-instance-domain required")
				}
				if subscriptionID != "" {
					if !strings.HasPrefix(subscriptionID, subscriptionsv1.EnterpriseSubscriptionIDPrefix) {
						return errors.Newf("subscription ID must start with %q", subscriptionsv1.EnterpriseSubscriptionIDPrefix)
					}
				}
				if subscriptionInstanceDomain != "" {
					var err error
					subscriptionInstanceDomain, err = subscriptionsv1.NormalizeInstanceDomain(subscriptionInstanceDomain)
					if err != nil {
						return errors.Wrap(err, "normalize instance domain")
					}
				}
				if len(roles) == 0 && !c.Bool("auto-approve") {
					var res string
					ok, err := std.PromptAndScan(std.Out, "No roles provided; all roles will be removed, are you sure? (y/N) ", &res)
					if err != nil {
						return err
					} else if !ok {
						return errors.New("response is required")
					}
					if res != "y" {
						return errors.New("aborting")
					}
				}
				pbRoles := make([]subscriptionsv1.Role, len(roles))
				for i, r := range roles {
					role, ok := subscriptionsv1.Role_value[r]
					if !ok {
						return errors.Newf("invalid role %q", r)
					}
					if role == 0 {
						return errors.Newf("invalid role %q", r)
					}
					pbRoles[i] = subscriptionsv1.Role(role)
				}

				samsClient, err := samsflags.NewClientFromFlags(c, scopes.Scopes{scopes.Profile})
				if err != nil {
					return errors.Wrap(err, "get SAMS client")
				}

				var subscriptionDebugText string
				if subscriptionID != "" {
					subscriptionDebugText = fmt.Sprintf("subscription %q", subscriptionID)
				} else {
					subscriptionDebugText = fmt.Sprintf("subscription with instance domain %q", subscriptionInstanceDomain)
				}
				std.Out.WriteSuggestionf("Assigning %d users roles [%s] to %s",
					len(members), strings.Join(roles, ", "), subscriptionDebugText)

				for _, member := range members {
					samsUserID, err := resolveUserReference(c.Context, samsClient.Users(), member)
					if err != nil {
						return errors.Wrap(err, "resolve SAMS user ID")
					}
					_, err = client.UpdateEnterpriseSubscriptionMembership(ctx, connect.NewRequest(&subscriptionsv1.UpdateEnterpriseSubscriptionMembershipRequest{
						Membership: &subscriptionsv1.EnterpriseSubscriptionMembership{
							SubscriptionId:      subscriptionID,
							InstanceDomain:      subscriptionInstanceDomain,
							MemberSamsAccountId: samsUserID,
							MemberRoles:         pbRoles,
						},
					}))
					if err != nil {
						return errors.Wrapf(err, "assign membership to user %q",
							samsUserID)
					}
				}
				std.Out.WriteSuccessf("Done!")
				return nil
			},
		}},
	}},
}
