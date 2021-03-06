package meta

import (
	"fmt"
	"path/filepath"
	"strconv"

	"github.com/pingcap-incubator/tiup-cluster/pkg/executor"
	"github.com/pingcap-incubator/tiup-cluster/pkg/template/scripts"
)

// DrainerComponent represents Drainer component.
type DrainerComponent struct{ *Specification }

// Name implements Component interface.
func (c *DrainerComponent) Name() string {
	return ComponentDrainer
}

// Instances implements Component interface.
func (c *DrainerComponent) Instances() []Instance {
	ins := make([]Instance, 0, len(c.Drainers))
	for _, s := range c.Drainers {
		s := s
		ins = append(ins, &DrainerInstance{instance{
			InstanceSpec: s,
			name:         c.Name(),
			host:         s.Host,
			port:         s.Port,
			sshp:         s.SSHPort,
			topo:         c.Specification,

			usedPorts: []int{
				s.Port,
			},
			usedDirs: []string{
				s.DeployDir,
				s.DataDir,
			},
			statusFn: func(_ ...string) string {
				url := fmt.Sprintf("http://%s:%d/status", s.Host, s.Port)
				return statusByURL(url)
			},
		}})
	}
	return ins
}

// DrainerInstance represent the Drainer instance.
type DrainerInstance struct {
	instance
}

// ScaleConfig deploy temporary config on scaling
func (i *DrainerInstance) ScaleConfig(e executor.TiOpsExecutor, b *Specification, cluster, user string, paths DirPaths) error {
	s := i.instance.topo
	defer func() {
		i.instance.topo = s
	}()
	i.instance.topo = b

	return i.InitConfig(e, cluster, user, paths)
}

// InitConfig implements Instance interface.
func (i *DrainerInstance) InitConfig(e executor.TiOpsExecutor, cluster, user string, paths DirPaths) error {
	if err := i.instance.InitConfig(e, cluster, user, paths); err != nil {
		return err
	}

	spec := i.InstanceSpec.(DrainerSpec)
	cfg := scripts.NewDrainerScript(
		i.GetHost()+":"+strconv.Itoa(i.GetPort()),
		i.GetHost(),
		paths.Deploy,
		paths.Data,
		paths.Log,
	).WithPort(spec.Port).WithNumaNode(spec.NumaNode).AppendEndpoints(i.instance.topo.Endpoints(user)...)

	cfg.WithCommitTs(spec.CommitTS)

	fp := filepath.Join(paths.Cache, fmt.Sprintf("run_drainer_%s_%d.sh", i.GetHost(), i.GetPort()))

	if err := cfg.ConfigToFile(fp); err != nil {
		return err
	}
	dst := filepath.Join(paths.Deploy, "scripts", "run_drainer.sh")
	if err := e.Transfer(fp, dst, false); err != nil {
		return err
	}

	if _, _, err := e.Execute("chmod +x "+dst, false); err != nil {
		return err
	}

	return i.mergeServerConfig(e, i.topo.ServerConfigs.Drainer, spec.Config, paths)
}
