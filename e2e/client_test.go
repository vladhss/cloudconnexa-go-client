package client

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/openvpn/cloudconnexa-go-client/v2/cloudconnexa"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func validateEnvVars(t *testing.T) {
	validateEnvVar(t, HostEnvVar)
	validateEnvVar(t, ClientIDEnvVar)
	validateEnvVar(t, ClientSecretEnvVar)
}

func validateEnvVar(t *testing.T, envVar string) {
	fmt.Println(os.Getenv(envVar))
	require.NotEmptyf(t, os.Getenv(envVar), "%s must be set", envVar)
}

const (
	HostEnvVar         = "OVPN_HOST"
	ClientIDEnvVar     = "CLOUDCONNEXA_CLIENT_ID"
	ClientSecretEnvVar = "CLOUDCONNEXA_CLIENT_SECRET"
)

func TestNewClient(t *testing.T) {
	c := setUpClient(t)
	assert.NotEmpty(t, c.Token)
}

func setUpClient(t *testing.T) *cloudconnexa.Client {
	validateEnvVars(t)
	var err error
	client, err := cloudconnexa.NewClient(os.Getenv(HostEnvVar), os.Getenv(ClientIDEnvVar), os.Getenv(ClientSecretEnvVar))
	require.NoError(t, err)
	return client
}

func TestListNetworks(t *testing.T) {
	c := setUpClient(t)
	response, err := c.Networks.GetByPage(0, 10)
	require.NoError(t, err)
	fmt.Printf("found %d networks\n", len(response.Content))
}

func TestListConnectors(t *testing.T) {
	c := setUpClient(t)
	response, err := c.NetworkConnectors.GetByPage(0, 10)
	require.NoError(t, err)
	fmt.Printf("found %d connectors\n", len(response.Content))
}

func TestCreateNetwork(t *testing.T) {
	c := setUpClient(t)
	timestamp := time.Now().Unix()
	testName := fmt.Sprintf("test-%d", timestamp)

	connector := cloudconnexa.NetworkConnector{
		Description: "test",
		Name:        testName,
		VpnRegionID: "it-mxp",
	}
	route := cloudconnexa.Route{
		Description: "test",
		Type:        "IP_V4",
		Subnet:      "10.189.253.64/30",
	}
	network := cloudconnexa.Network{
		Description:    "test",
		Egress:         false,
		Name:           testName,
		InternetAccess: cloudconnexa.InternetAccessSplitTunnelOn,
		Connectors:     []cloudconnexa.NetworkConnector{connector},
	}
	response, err := c.Networks.Create(network)
	require.NoError(t, err)
	fmt.Printf("created %s network\n", response.ID)
	test, err := c.Routes.Create(response.ID, route)
	require.NoError(t, err)
	fmt.Printf("created %s route\n", test.ID)
	serviceConfig := cloudconnexa.IPServiceConfig{
		ServiceTypes: []string{"ANY"},
	}
	ipServiceRoute := cloudconnexa.IPServiceRoute{
		Description: "test",
		Value:       "10.189.253.64/30",
	}
	service := cloudconnexa.IPService{
		Name:            testName,
		Description:     "test",
		NetworkItemID:   response.ID,
		Type:            "IP_SOURCE",
		NetworkItemType: "NETWORK",
		Config:          &serviceConfig,
		Routes:          []*cloudconnexa.IPServiceRoute{&ipServiceRoute},
	}
	s, err := c.NetworkIPServices.Create(&service)
	require.NoError(t, err)
	fmt.Printf("created %s service\n", s.ID)
	err = c.Networks.Delete(response.ID)
	require.NoError(t, err)
	fmt.Printf("deleted %s network\n", response.ID)
}
