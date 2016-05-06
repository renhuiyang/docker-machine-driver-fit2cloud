package main

import (
	"github.com/docker/machine/libmachine/drivers/plugin"
	//"github.com/docker/machine/libmachine/ssh"
	"renh.yang/docker-machine-driver-fit2cloud/fit2cloud"
)

func main() {
	//ssh.SetDefaultClient(ssh.Native)
	plugin.RegisterDriver(fit2cloud.NewDriver("", ""))
}
