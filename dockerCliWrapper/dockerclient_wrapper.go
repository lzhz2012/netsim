package dockerCliWrapper

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	network "github.com/docker/docker/api/types/network"
	dockerclient "github.com/docker/docker/client"
	log "github.com/sirupsen/logrus"
)

type DockerCliCfg struct {
	//以下是server Damemon相关
	DockerSeverIp    string
	DockerSeverPort  string
	DockerApiVersion string
}

type BuildCfg struct {
	//下面两个是build相关
	ImageName   string
	TarFile     string
	ProjectName string
}

type ContainerCfg struct {
	Image         string   `json:"image" yml:"image"`
	ContainerName string   `json:"container_name" yml:"container_name"`
	Hostname      string   `json:"hostname" yml:"hostname"`
	Networks      []string `json:"networks" yml:"networks"`
	Volumes       []string `json:"volumes" yml:"volumes"`
	Environment   []string `json:"environment" yml:"environment"`
	Restart       string   `json:"restart" yml:"restart"`
}

type PushCfg struct {
	RegistryUser     string `json:"registryUsr" yml:"registryUsr"`
	RegistryPassword string `json:"registryPass" yml:"registryPass"`
	RegistryIP       string `json:"registryIP" yml:"registryIP"`
	RegistryPort     string `json:"registryPort" yml:"registryPort"`
	Image            string
}

type DockerCliWrapper struct {
	DockerCli    *dockerclient.Client
	DockerCliCfg DockerCliCfg
}

func CheckContainerConfig(cfg *ContainerCfg) error {
	if cfg == nil || cfg.Image == "" {
		return errors.New("container Image Name should be not empty")
	}
	return nil
}

var cli DockerCliWrapper

func NewClient(dockerCliCfg *DockerCliCfg) (*DockerCliWrapper, error) {

	//cli.DockerCli, err = dockerclient.NewClientWithOpts(dockerclient.FromEnv, dockerclient.WithVersion(dockerCliCfg.DockerApiVersion)) //仅1.42版本支持
	//使用定制化docker Api中Ip和Port不知道怎么填写
	//dockerHost := "tcp://" + dockerCliCfg.DockerSeverIp + ":" + dockerCliCfg.DockerSeverPort
	//cli.DockerCli, err = dockerclient.NewClient(dockerHost, dockerCliCfg.DockerApiVersion, nil, map[string]string{"Content-type": "application/x-tar"})
	os.Setenv("DOCKER_API_VERSION", dockerCliCfg.DockerApiVersion)
	dockercli, err := dockerclient.NewEnvClient()
	if err != nil {
		return nil, err
	}
	//dockercli.SetCustomHTTPHeaders(map[string]string{"Content-type": "application/x-tar"})
	cli.DockerCli = dockercli
	return &cli, nil
}

func (cli *DockerCliWrapper) PushImage(cfg *PushCfg) error {

	if cli == nil {
		log.Printf("Push image error: client null pointer")
		return fmt.Errorf("client null pointer")
	}
	authConfig := types.AuthConfig{
		Username:      cfg.RegistryUser,
		Password:      cfg.RegistryPassword,
		ServerAddress: cfg.RegistryIP + cfg.RegistryPort,
	}
	encodedJSON, err := json.Marshal(authConfig)
	if err != nil {
		return err
	}
	authStr := base64.URLEncoding.EncodeToString(encodedJSON)
	imagPushCfg := types.ImagePushOptions{
		RegistryAuth: authStr,
	}

	out, err := cli.DockerCli.ImagePush(context.Background(), cfg.Image, imagPushCfg)
	if err != nil {
		return err
	}

	body, err := ioutil.ReadAll(out)
	if err != nil {
		return err
	}
	log.Printf("Push docker image output: %v", string(body))

	if strings.Contains(string(body), "error") {
		return fmt.Errorf("push image to docker error")
	}

	log.Printf("Push docker image to registry: %+v success", cfg)

	return nil
}

func (cli *DockerCliWrapper) BuildImage(cfg *BuildCfg) error {
	dockerBuildContext, err := os.Open(cfg.TarFile)
	if err != nil {
		return err
	}
	defer dockerBuildContext.Close()

	buildOptions := types.ImageBuildOptions{
		Dockerfile: "Dockerfile", // optional, is the default
		Tags:       []string{cfg.ImageName},
		Labels: map[string]string{
			"project": cfg.ProjectName,
		},
	}

	output, err := cli.DockerCli.ImageBuild(context.Background(), dockerBuildContext, buildOptions)
	if err != nil {
		return err
	}

	body, err := ioutil.ReadAll(output.Body)
	if err != nil {
		return err
	}
	var content string
	//err = json.Unmarshal(body, content)
	content = string(body)
	log.Printf("Build resource image output: %s", content)

	if !strings.Contains(string(body), "Successfully built") {
		return fmt.Errorf("build image to docker error")
	}

	return nil
}

func (cli *DockerCliWrapper) StartContainer(cfg *ContainerCfg) error {

	if err := CheckContainerConfig(cfg); err != nil {
		log.Printf("check container config falied, err: %s", err)
		return err
	}
	dockercli := cli.DockerCli
	hostConfig := &container.HostConfig{
		LogConfig: container.LogConfig{
			Type:   "json-file",
			Config: map[string]string{},
		},
	}
	if len(cfg.Volumes) > 0 {
		hostConfig.Binds = cfg.Volumes // Binds is the actual "-v" volume.
	}
	if len(cfg.Restart) > 0 {
		hostConfig.RestartPolicy.Name = cfg.Restart
	}

	// Look for Shuffle network and set it
	networkConfig := &network.NetworkingConfig{}
	if len(cfg.Networks) > 0 {
		// 目前仅遍历第一个网络
		ctx := context.Background()
		networkResources, err := dockercli.NetworkList(ctx, types.NetworkListOptions{})
		if err != nil {
			log.Printf("list docker networks failed, err: %s", err)
		}
		found := false
		for _, netRes := range networkResources {
			if strings.Contains(netRes.Name, cfg.Networks[0]) {
				found = true
			}

		}
		if !found {
			networkOptions := types.NetworkCreate{
				Driver: "bridge",
			}
			dockercli.NetworkCreate(ctx, cfg.Networks[0], networkOptions)
		}

		networkConfig = &network.NetworkingConfig{
			EndpointsConfig: map[string]*network.EndpointSettings{
				cfg.Networks[0]: {
					NetworkID: cfg.Networks[0],
				},
			},
		}
	}

	config := &container.Config{
		Image: cfg.Image,
	}
	if len(cfg.Environment) > 0 {
		config.Env = cfg.Environment
	}

	//如果容器已经存在先强制移除
	removeOptions := types.ContainerRemoveOptions{
		RemoveVolumes: true,
		Force:         true,
	}
	if err := dockercli.ContainerRemove(context.Background(), cfg.ContainerName, removeOptions); err != nil {
		//log.Printf("Unable to remove container: %s", err)
	}

	cont, err := dockercli.ContainerCreate(
		context.Background(),
		config,
		hostConfig,
		networkConfig,
		cfg.ContainerName,
	)

	if err != nil {
		log.Println(err)
		return err
	}

	// FIXME：这里是否可以开启协程，毕竟开启容器的等待时间比较长，但是我们又需要这样的结果（用channel？）
	err = dockercli.ContainerStart(context.Background(), cont.ID, types.ContainerStartOptions{})
	if err != nil {
		log.Printf("Failed to start container in environment %s: %s", cfg.Environment, err)
		return err
	} else {
		log.Printf("Container %s was created under environment %s", cont.ID, cfg.Environment)
	}

	return nil
}

func (cli *DockerCliWrapper) RemoveContainer(containername string) error {
	var err error

	dockercli := cli.DockerCli
	if dockercli == nil {
		log.Printf("dockerclient is nil")
		return errors.New("docker client is nil")
	}

	ctx := context.Background()
	if err := dockercli.ContainerStop(ctx, containername, nil); err != nil {
		log.Printf("Unable to stop container %s - running removal anyway, just in case: %s", containername, err)
	}

	removeOptions := types.ContainerRemoveOptions{
		RemoveVolumes: true,
		Force:         true,
	}

	if err := dockercli.ContainerRemove(ctx, containername, removeOptions); err != nil {
		log.Printf("Unable to remove container: %s", err)
	}

	return err
}

//FIXME:目前仅仅做了在单一机器上的判断，后面分布式的时候
func (cli *DockerCliWrapper) IsContainerRunning(name string) bool {
	name = "/" + name
	listOps := types.ContainerListOptions{All: true}
	containers, err := cli.DockerCli.ContainerList(context.Background(), listOps)
	if err != nil {
		log.Error(fmt.Sprintf("Unable to list docker containers: %s", err))
	}
	for _, container := range containers {
		for _, item := range container.Names {
			if (item == name) && (container.State == "running") {
				return true
			}
		}
	}

	return false
}
