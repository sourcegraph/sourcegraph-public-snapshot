package database

import (
	"context"
	"net/netip"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type ServiceArgs struct {
	// We ignore the IP address for JSON marshalling because we extract it from the
	// request and not from the JSON payload.
	IP netip.Addr `json:"-"`

	// Required
	// Valid ports are >0.
	Port uint16 `json:"port"`
	// The self-reported hostname of the server.
	Hostname string `json:"hostname"`
}

type ServicesStore interface {
	Register(ctx context.Context, service string, args ServiceArgs) (string, error)
	Renew(ctx context.Context, service, id string) error
	Deregister(ctx context.Context, service, id string) error

	GetByService(ctx context.Context, service string) ([]ServiceArgs, error)
	Invalidate(ctx context.Context, age time.Duration) error
}

type servicesStore struct {
	logger log.Logger
	*basestore.Store
}

func ServicesWith(logger log.Logger, other basestore.ShareableStore) ServicesStore {
	return &servicesStore{
		logger: logger,
		Store:  basestore.NewWithHandle(other.Handle()),
	}
}

var _ ServicesStore = (*servicesStore)(nil)

const registerFmtStr = `
-- source: /internal/database/service_registry.go:Register
INSERT INTO service_registry (service, ip, port, hostname)
VALUES (%s, %s, %d, %s)`

func (s servicesStore) Register(ctx context.Context, service string, args ServiceArgs) (string, error) {
	err := s.Exec(ctx, sqlf.Sprintf(registerFmtStr, service, args.IP.String(), args.Port, args.Hostname))
	if err != nil {
		return "", err
	}

	// We use the combination of IP and port as instanceID. However, the ID is
	// arbitrary as long as it is unique for each instance. By using IP and
	// port we avoid storing an extra string field in the DB. The caller should
	// treat the returned ID as an opaque string.
	return netip.AddrPortFrom(args.IP, args.Port).String(), nil
}

const renewFmtStr = `
-- source: /internal/database/service_registry.go:Renew
UPDATE service_registry
SET last_heartbeat = now()
WHERE service = %s
AND ip = %s
AND port = %d`

func (s servicesStore) Renew(ctx context.Context, service, id string) error {
	return s.renewOrDeregister(ctx, renewFmtStr, service, id)
}

const deregisterFmtStr = `
-- source: /internal/database/service_registry.go:Deregister
DELETE FROM service_registry
WHERE service = %s
AND ip = %s
AND port = %d`

func (s servicesStore) Deregister(ctx context.Context, service, id string) error {
	return s.renewOrDeregister(ctx, deregisterFmtStr, service, id)
}

func (s servicesStore) renewOrDeregister(ctx context.Context, queryStr, service, id string) error {
	addrPort, err := netip.ParseAddrPort(id)
	if err != nil {
		return errors.Wrapf(err, "id=%q", id)
	}

	res, err := s.ExecResult(ctx, sqlf.Sprintf(queryStr, service, addrPort.Addr().String(), addrPort.Port()))
	if err != nil {
		return err
	}

	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return NotFoundError{errors.New("unknown service or id")}
	}
	return nil
}

type NotFoundError struct {
	error
}

func (e NotFoundError) NotFound() bool {
	return true
}

const getByServiceFmtStr = `
-- source: /internal/database/service_registry.go:GetByService
SELECT ip, port, hostname FROM service_registry where service = %s`

func (s servicesStore) GetByService(ctx context.Context, service string) (instances []ServiceArgs, err error) {
	rows, err := s.Query(ctx, sqlf.Sprintf(getByServiceFmtStr, service))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ipStr := ""
	for rows.Next() {
		instance := ServiceArgs{}
		if err := rows.Scan(
			&ipStr,
			&instance.Port,
			&instance.Hostname,
		); err != nil {
			return nil, err
		}
		instance.IP, err = netip.ParseAddr(ipStr)
		if err != nil {
			return nil, err
		}
		instances = append(instances, instance)
	}
	return
}

const invalidateFmtStr = `
-- source: /internal/database/service_registry.go:Invalidate
DELETE from service_registry
WHERE last_heartbeat < (NOW() - (%s * '1 second'::interval));`

func (s servicesStore) Invalidate(ctx context.Context, age time.Duration) error {
	return s.Exec(ctx, sqlf.Sprintf(invalidateFmtStr, int(age/time.Second)))
}
