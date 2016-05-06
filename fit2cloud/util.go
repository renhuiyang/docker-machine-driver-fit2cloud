package fit2cloud

import (
	"errors"
	"github.com/docker/machine/libmachine/state"
	fit2cloud "renh.yang/fit2cloud-go-sdk"
	"github.com/docker/machine/libmachine/log"
	"renh.yang/fit2cloud-go-sdk/model"
)

func (d *Driver) newAzureClient() (*fit2cloud.Fit2CloudClient, error) {
	return fit2cloud.NewClient(d.Consumer, d.Secret, d.Endpoint), nil
}

func (d *Driver) getClusterId() (int64, error) {
	client, err := d.newAzureClient()
	if err != nil {
		return 0, err
	}

	clusters, err := client.GetClusters()
	if err != nil {
		return 0, err
	}

	clusterId := int64(-1)
	for _, cluster := range clusters {
		if cluster.Name == d.Cluster {
			clusterId = cluster.Id
			break
		}
	}

	if clusterId < 0 {
		return clusterId, errors.New(d.Cluster + " not been found!")
	}
	return clusterId, nil
}

func (d *Driver) getClusterRoleId() (int64, error) {
	client, err := d.newAzureClient()
	if err != nil {
		return 0, err
	}

	clusterRoleId := int64(-1)
	clusterRoles, err := client.GetClusterRoles(d.CLusterId)
	if err != nil {
		return clusterRoleId, err
	}

	for _, clusterRole := range clusterRoles {
		if clusterRole.Name == d.ClusterRole {
			clusterRoleId = clusterRole.Id
			break
		}
	}

	if clusterRoleId < 0 {
		return clusterRoleId, errors.New(d.ClusterRole + " not been found in " + d.Cluster)
	}

	return clusterRoleId, nil
}
func (d *Driver) getTemplateId() (int64, error) {
	client, err := d.newAzureClient()
	if err != nil {
		return 0, err
	}

	cnfId := int64(-1)
	if cnfs, err := client.GetLaunchconfiguration(0); err != nil {
		return cnfId, err
	} else {
		for _, cnf := range cnfs {
			if cnf.Name == d.Templatename {
				cnfId = cnf.Id
				break
			}
		}
	}
	if cnfId < 0 {
		return cnfId, errors.New(d.Templatename + " not been found!")
	}

	return cnfId, nil
}

func (d *Driver) getServerId(servername string) (int64, error) {
	client, err := d.newAzureClient()
	if err != nil {
		return 0, err
	}

	serverId := int64(-1)
	if cnfs, err := client.GetServers(d.CLusterId, d.ClusterRoleId, "", "", -1, -1, false); err != nil {
		return serverId, err
	} else {
		for _, cnf := range cnfs {
			if cnf.Name == servername {
				serverId = cnf.Id
				break
			}
		}
	}
	if serverId < 0 {
		return serverId, errors.New(d.MachineName + " not been found!")
	}
	d.ServerId = serverId

	return serverId, nil
}

func (d *Driver) lanuchVM() error {
	client, err := d.newAzureClient()
	if err != nil {
		return err
	}

	server, err := client.LaunchServerAsync(d.CLusterId, d.ClusterRoleId, d.TemplateId)
	if err != nil {
		return err
	}
	d.MachineName = server.Name
	d.ServerId = server.Id
	return err
}

func (d *Driver) removeVm() error {
	client, err := d.newAzureClient()
	if err != nil {
		return err
	}
	serverId, err := d.getServerId(d.MachineName)

	if err != nil {
		return err
	}
	if client.TerminateServer(serverId) {
		return nil
	} else {
		return errors.New("Terminate " + d.MachineName + " failed!")
	}
}

func (d *Driver) getServer() (model.Server, error) {
	client, err := d.newAzureClient()
	if err != nil {
		return model.Server{}, err
	}
	log.Debugf("ServerID:%v ServerName:%v\n",d.ServerId,d.MachineName)
	return client.GetServer(d.ServerId)
}

func (d *Driver) startServer() error {
	client, err := d.newAzureClient()
	if err != nil {
		return err
	}
	_, err = client.StartServer(d.ServerId)
	return err
}

func (d *Driver) stopServer() error {
	client, err := d.newAzureClient()
	if err != nil {
		return err
	}
	if client.StopServer(d.ServerId) {
		return nil
	}
	return errors.New("Stop server failed!")
}

func (d *Driver) stateForFit2cloudStatus(status string) state.State {
	m := map[string]state.State{
		"Starting":     state.Starting,
		"Running":      state.Running,
		"Stopping":     state.Stopping,
		"Stopped":      state.Stopped,
		"Deallocating": state.Stopping,
		"Deallocated":  state.Stopped,
		"Unknown":      state.None,
	}
	if v, ok := m[status]; ok {
		return v
	}
	return state.None
}
