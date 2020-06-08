package config

import (
	"context"
	"io/ioutil"
	"strconv"
	"sync"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/k0kubun/pp"
	"github.com/rahveiz/topomate/utils"

	"gopkg.in/yaml.v2"
)

type Router struct {
	Hostname      string
	ContainerName string
}
type AutonomousSystem struct {
	ASN        int        `yaml:"asn"`
	NumRouters int        `yaml:"routers"`
	IGP        string     `yaml:"igp"`
	Routers    []Router   `yaml:"-"`
	Links      LinkModule `yaml:"links"`
}
type BaseConfig struct {
	As []*AutonomousSystem `yaml:"autonomous_systems"`
}

func ReadConfig(path string) *BaseConfig {
	conf := &BaseConfig{}
	data, err := ioutil.ReadFile(path)
	utils.Check(err)
	err = yaml.Unmarshal(data, conf)
	utils.Check(err)

	// Generate routers
	for _, k := range conf.As {
		k.Routers = make([]Router, k.NumRouters)
		for i := 0; i < k.NumRouters; i++ {
			host := "R" + strconv.Itoa(i+1)
			k.Routers[i] = Router{
				Hostname:      host,
				ContainerName: "AS" + strconv.Itoa(k.ASN) + "-" + host,
			}
		}
	}

	return conf
}

func (c *BaseConfig) Print() {
	for _, v := range c.As {
		pp.Print(*v)
	}
}

func (c *BaseConfig) StartAll() {
	var wg sync.WaitGroup
	for _, v := range c.As {
		wg.Add(v.NumRouters)
		for i := 0; i < len(v.Routers); i++ {
			go v.Routers[i].StartContainer(&wg)
		}
	}
	wg.Wait()
}

func (c *BaseConfig) StopAll() {
	var wg sync.WaitGroup
	for _, v := range c.As {
		wg.Add(v.NumRouters)
		for i := 0; i < len(v.Routers); i++ {
			go v.Routers[i].StopContainer(&wg)
		}
	}
	wg.Wait()
}

func (r *Router) StartContainer(wg *sync.WaitGroup) {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	utils.Check(err)

	if wg != nil {
		defer wg.Done()
	}

	// Check if container already exists
	var containerID string
	flt := filters.NewArgs(filters.Arg("name", r.ContainerName))
	li, err := cli.ContainerList(ctx, types.ContainerListOptions{
		All:     true,
		Filters: flt,
	})
	if len(li) == 0 { // container does not exist yet
		resp, err := cli.ContainerCreate(ctx, &container.Config{
			Image: "topomate-router",
		}, nil, nil, nil, r.ContainerName)
		utils.Check(err)
		containerID = resp.ID
	} else { // container exists
		containerID = li[0].ID
	}

	// Start container
	if err := cli.ContainerStart(ctx, containerID, types.ContainerStartOptions{}); err != nil {
		panic(err)
	}

}

func (r *Router) StopContainer(wg *sync.WaitGroup) {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	utils.Check(err)

	if wg != nil {
		defer wg.Done()
	}

	if err := cli.ContainerStop(ctx, r.ContainerName, nil); err != nil {
		panic(err)
	}
}

// func ()