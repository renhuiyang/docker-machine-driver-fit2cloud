package fit2cloud

import (
	"net"

	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/mcnflag"
	"github.com/docker/machine/libmachine/state"
	"net/url"
	"strconv"
	"time"
)

const (
	defaulttemplatename    = "mytemplatename"
	defaultClusterName     = "rancherHostCluster"
	defaultCLusterRoleName = "rancherHostClusterRole"
)

type Driver struct {
	*drivers.BaseDriver
	Templatename  string
	Consumer      string
	Secret        string
	Endpoint      string
	Cluster       string
	ClusterRole   string
	CLusterId     int64
	ClusterRoleId int64
	TemplateId    int64
	ServerId      int64
	UserPassword  string
}

func NewDriver(hostName, storePath string) *Driver {
	return &Driver{
		BaseDriver: &drivers.BaseDriver{
			MachineName: hostName,
			StorePath:   storePath,
		},
		Templatename: defaulttemplatename,
		Cluster:      defaultClusterName,
		ClusterRole:  defaultCLusterRoleName,
	}
}

func (d *Driver) GetCreateFlags() []mcnflag.Flag {
	return []mcnflag.Flag{
		mcnflag.StringFlag{
			EnvVar: "FIT2CLOUD_TEMPLATE",
			Name:   "fit2cloud-template",
			Usage:  "select the template will be used",
			Value:  defaulttemplatename,
		},
		mcnflag.StringFlag{
			EnvVar: "CONSUMER_KEY",
			Name:   "consumer-key",
			Usage:  "consumer key",
		},
		mcnflag.StringFlag{
			EnvVar: "SECRET_KEY",
			Name:   "secret-key",
			Usage:  "secret key",
		},
		mcnflag.StringFlag{
			EnvVar: "FIT2CLOUD_ENDPOINT",
			Name:   "fit2cloud-endpoint",
			Usage:  "fit2cloud endpoint",
		},
		mcnflag.StringFlag{
			EnvVar: "FIT2CLOUD_CLUSTER",
			Name:   "fit2cloud-cluster",
			Usage:  "fit2cloud cluster",
			Value:  defaultClusterName,
		},
		mcnflag.StringFlag{
			EnvVar: "FIT2CLOUD_CLUSTERROLE",
			Name:   "fit2cloud-cluster-role",
			Usage:  "fit2cloud cluster role",
			Value:  defaultCLusterRoleName,
		},
	}
}

func (d *Driver) GetMachineName() string {
	return d.MachineName
}

// DriverName returns the name of the driver.
func (d *Driver) DriverName() string { return "fit2cloud" }

//func (d *Driver)GetSSHHostname()(string,error){
//	return d.GetIP()
//}
//
//func (d *Driver)GetSSHKeyPath()string{
//	return d.ResolveStorePath("id_rsa")
//}

func (d *Driver) SetConfigFromFlags(flags drivers.DriverOptions) error {
	d.Templatename = flags.String("fit2cloud-template")
	d.Consumer = flags.String("consumer-key")
	d.Secret = flags.String("secret-key")
	d.Endpoint = flags.String("fit2cloud-endpoint")
	d.Cluster = flags.String("fit2cloud-cluster")
	d.ClusterRole = flags.String("fit2cloud-cluster-role")

	d.SSHUser = "hna"
	d.UserPassword = "Hna!1qwe"

	log.Debugf("F2CDriver:%v\n", *d)
	return nil
}

func (d *Driver) PreCreateCheck() error {
	clusterId, err := d.getClusterId()
	if err != nil {
		return err
	}
	d.CLusterId = clusterId
	clusterRoleId, err := d.getClusterRoleId()
	if err != nil {
		return err
	}

	d.ClusterRoleId = clusterRoleId

	templateId, err := d.getTemplateId()
	if err != nil {
		return err
	}

	d.TemplateId = templateId
	return nil
}

func (d *Driver) Create() error {
	if err := d.lanuchVM(); err != nil {
		return err
	}
	for true {
		if status, err := d.GetState(); err != nil {
			return err
		} else {
			if status != state.Starting && status != state.None {
				return nil
			}
			time.Sleep(1 * time.Minute)
		}
	}
	return nil
}

func (d *Driver) Remove() error {
	return d.removeVm()
}

// GetIP returns public IP address or hostname of the machine instance.
func (d *Driver) GetIP() (string, error) {
	server, err := d.getServer()
	if err != nil {
		return "", nil
	}

	return server.RemoteIP, nil
}

// GetSSHHostname returns an IP address or hostname for the machine instance.
func (d *Driver) GetSSHHostname() (string, error) {
	return d.GetIP()
}

// GetURL returns a socket address to connect to Docker engine of the machine
// instance.
func (d *Driver) GetURL() (string, error) {
	ip, err := d.GetIP()
	if err != nil {
		return "", err
	}
	u := (&url.URL{
		Scheme: "tcp",
		Host:   net.JoinHostPort(ip, strconv.Itoa(2376)),
	}).String()
	log.Debugf("Machine URL is resolved to: %s", u)
	return u, nil
}

// GetState returns the state of the virtual machine role instance.
func (d *Driver) GetState() (state.State, error) {
	server, err := d.getServer()
	if err != nil {
		return state.Error, err
	}
	return d.stateForFit2cloudStatus(server.VmStatus), nil
}

// Start issues a power on for the virtual machine instance.
func (d *Driver) Start() error {
	return d.startServer()
}

// Stop issues a power off for the virtual machine instance.
func (d *Driver) Stop() error {
	return d.stopServer()
}

// Restart reboots the virtual machine instance.
func (d *Driver) Restart() error {
	if err := d.Start(); err != nil {
		return err
	}

	if err := d.Stop(); err != nil {
		return err
	}
	if err := d.Start(); err != nil {
		return err
	}
	return nil
}

// Kill stops the virtual machine role instance.
func (d *Driver) Kill() error {
	return d.Remove()
}

//add for passwd ssh login -yrh
func (d *Driver) GetPassword() string {
	return d.UserPassword
}

func (d *Driver)GetSSHKeyPath() string{
	return ""
}
