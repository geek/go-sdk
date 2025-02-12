package dsdk

import (
	"context"
	"fmt"
	"net/http"

	udc "github.com/Datera/go-udc/pkg/udc"
	uuid "github.com/google/uuid"
)

const (
	VERSION         = "1.1.5"
	VERSION_HISTORY = `
		1.1.0 -- Revamped SDK to new directory structure, switched to using grequests and added UDC support
		1.1.1 -- Added LDAP server support
		1.1.2 -- Added logs upload, template override
		1.1.3 -- Support for Go modules
		1.1.4 -- AppInstance AppTemplate datastructure bugfix
		1.1.5 -- HTTP 503 Retry and Connection Retry support
	`
)

type SDK struct {
	conf                 *udc.UDC
	Conn                 *ApiConnection
	Ctxt                 context.Context
	AccessNetworkIpPools *AccessNetworkIpPools
	AppInstances         *AppInstances
	AppTemplates         *AppTemplates
	Initiators           *Initiators
	InitiatorGroups      *InitiatorGroups
	LogsUpload           *LogsUpload
	HWMetrics            *HWMetrics
	IOMetrics            *IOMetrics
	PlacementPolicies    *PlacementPolicies
	RemoteProvider       *RemoteProviders
	StorageNodes         *StorageNodes
	StoragePools         *StoragePools
	System               *System
	SystemEvents         *SystemEvents
	Tenants              *Tenants
	UserData             *UserDatas
}

func NewSDK(c *udc.UDC, secure bool) (*SDK, error) {
	return NewSDKWithHTTPClient(c, secure, nil)
}

func NewSDKWithHTTPClient(c *udc.UDC, secure bool, client *http.Client) (*SDK, error) {
	var err error
	if c == nil {
		c, err = udc.GetConfig()
		if err != nil {
			Log().Error(err)
			return nil, err
		}
	}
	conn := NewApiConnectionWithHTTPClient(c, secure, client)
	return &SDK{
		conf:                 c,
		Conn:                 conn,
		AccessNetworkIpPools: newAccessNetworkIpPools("/"),
		AppInstances:         newAppInstances("/"),
		AppTemplates:         newAppTemplates("/"),
		Initiators:           newInitiators("/"),
		InitiatorGroups:      newInitiatorGroups("/"),
		LogsUpload:           newLogsUpload("/"),
		HWMetrics:            newHWMetrics("/"),
		IOMetrics:            newIOMetrics("/"),
		PlacementPolicies:    newPlacementPolicies("/"),
		RemoteProvider:       newRemoteProviders("/"),
		StorageNodes:         newStorageNodes("/"),
		StoragePools:         newStoragePools("/"),
		System:               newSystem("/"),
		SystemEvents:         newSystemEvents("/"),
		Tenants:              newTenants("/"),
		UserData:             newUserDatas("/"),
	}, nil
}

func (c SDK) SetDriver(d string) {
	DateraDriver = d
}

func (c SDK) WithContext(ctxt context.Context) context.Context {
	return context.WithValue(ctxt, "conn", c.Conn)
}

func (c SDK) NewContext() context.Context {
	ctxt := context.WithValue(context.Background(), "conn", c.Conn)
	ctxt = context.WithValue(ctxt, "tid", uuid.Must(uuid.NewRandom()).String())
	return ctxt
}

func (c SDK) GetDateraVersion() (string, error) {
	sys, apierr, err := c.System.Get(&SystemGetRequest{
		Ctxt: context.WithValue(c.NewContext(), "quiet", true),
	})
	if err != nil {
		return "", err
	}
	if apierr != nil {
		return "", fmt.Errorf("ApiError: %s", Pretty(apierr))
	}
	return sys.SwVersion, nil
}

// Cleans AppInstances, AppTemplates, StorageInstances, Initiators and InitiatorGroups under
// the currently configured tenant
func (c SDK) HealthCheck() error {
	sns, apierr, err := c.StorageNodes.List(&StorageNodesListRequest{
		Ctxt: context.WithValue(c.NewContext(), "quiet", true),
	})
	if err != nil {
		return err
	}
	if apierr != nil {
		return fmt.Errorf("ApiError: %s", Pretty(apierr))
	}
	Log().Debugf("Connected to cluster: %s with tenant %s.", c.conf.MgmtIp, c.conf.Tenant)
	for _, sn := range sns {
		Log().Debugf("Found Storage Node: %s", sn.Uuid)
	}
	return nil
}
