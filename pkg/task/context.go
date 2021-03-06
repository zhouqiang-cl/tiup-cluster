package task

import (
	"github.com/pingcap-incubator/tiup-cluster/pkg/executor"
	"github.com/pingcap-incubator/tiup-cluster/pkg/meta"
	"github.com/pingcap/errors"
)

// SetSSHKeySet set ssh key set.
func (ctx *Context) SetSSHKeySet(privateKeyPath string, publicKeyPath string) error {
	ctx.PrivateKeyPath = privateKeyPath
	ctx.PublicKeyPath = publicKeyPath
	return nil
}

// SetClusterSSH set cluster user ssh executor in context.
func (ctx *Context) SetClusterSSH(topo *meta.Specification, deployUser string) error {
	if len(ctx.PrivateKeyPath) == 0 {
		return errors.Errorf("context has no PrivateKeyPath")
	}

	for _, com := range topo.ComponentsByStartOrder() {
		for _, in := range com.Instances() {
			cf := executor.SSHConfig{
				Host:    in.GetHost(),
				KeyFile: ctx.PrivateKeyPath,
				User:    deployUser,
			}

			e := executor.NewSSHExecutor(cf)
			ctx.SetExecutor(in.GetHost(), e)
		}
	}

	return nil
}
