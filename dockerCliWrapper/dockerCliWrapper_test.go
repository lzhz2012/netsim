package dockerCliWrapper

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/lzhz2012/netsim/utils"
	log "github.com/sirupsen/logrus"
)

func TestStartContainer(t *testing.T) {
	dockerCliCfg := &DockerCliCfg{
		DockerApiVersion: "1.40", DockerSeverIp: "localhost", DockerSeverPort: "5000",
	}
	cli, err := NewClient(dockerCliCfg)
	if err != nil {
		log.Error(fmt.Sprintf("Unable to create docker client: %s", err))
	}
	// Networks      []string `json:"networks" yml:"networks"`
	// Volumes       []string `json:"volumes" yml:"volumes"`
	// Environment   []string `json:"environment" yml:"environment"`
	// 测试并行启动容器
	cfg := &ContainerCfg{
		Image:    "hello-world:latest",
		Networks: []string{"shuffle_shuffle"},
	}
	for i := 0; i < 3; i++ {
		cfg.ContainerName = "hello-word" + strconv.Itoa(i)
		go cli.StartContainer(cfg)
	}
	time.Sleep(30 * time.Second)
}

var remoteRegistryField string = "10.1.5.224:5000/"
var shuffleField string = "frikky/shuffle"
var newTagBase string = "10.1.5.224:5000/shuffle/"

func PushImagesForShuffle() error {
	ctx := context.Background()
	dockerCliCfg := &DockerCliCfg{DockerApiVersion: "1.40"}
	cli, _ := NewClient(dockerCliCfg)
	dockercli := cli.DockerCli
	images, err := dockercli.ImageList(ctx, types.ImageListOptions{})
	if err != nil {
		log.Error(fmt.Sprintf("Unable to list docker images: %s", err))
		return err
	}

	cfg := &PushCfg{
		RegistryUser: "registry", RegistryPassword: "registry",
		RegistryIP: "10.1.5.224", RegistryPort: "5000",
		Image: "",
	}

	// eg: host:port/imageName:version
	var pushfailedImages []string
	var tagfailedImages []string
	for _, img := range images {
		log.Printf("Image info:%+v", img)
		imageVersion := "latest" // eg: "1.0.0"
		for _, tag := range img.RepoTags {
			//if strings.Contains(tag, remoteRegistryField) {
			if strings.Contains(tag, shuffleField) {
				countSplit := strings.Split(tag, ":")
				lastIdx := 0
				if len(countSplit)-1 >= 0 {
					lastIdx = len(countSplit) - 1
				}
				countSplit1 := strings.Split(countSplit[lastIdx], "_")
				if len(countSplit1) > 1 {
					imageVersion = countSplit1[1]
				}
				// docker标签冒号前不支持大写，冒号后支持大写（version支持）
				//newTag := newTagBase + strings.ToLower(countSplit1[0]) + ":" + imageVersion
				newTag := newTagBase + countSplit1[0] + ":" + imageVersion
				//可以重复打tag
				if err := dockercli.ImageTag(ctx, tag, newTag); err != nil {
					log.Printf("tag image failed, tag:%s, err:%s", tag, err)
					//return err
					tagfailedImages = append(tagfailedImages, tag)
					continue
				}
				fmt.Println("find iamge, old tag", tag, "new tag:", newTag)
				cfg.Image = newTag
				if err := cli.PushImage(cfg); err != nil {
					log.Printf("Push docker image:%s to registry failed, err:%s", cfg.Image, err)
					pushfailedImages = append(pushfailedImages, cfg.Image)
				}
			}
		}

	}
	if len(pushfailedImages) > 0 {
		log.Printf("some docker images:%s push to registry failed", pushfailedImages)
	}
	if len(tagfailedImages) > 0 {
		log.Printf("some docker images:%s taged failed", tagfailedImages)
	}
	return nil
}

func TestPushImages(t *testing.T) {
	_ = PushImagesForShuffle()
}

func TestRemoveContainer(t *testing.T) {
	dockerCliCfg := &DockerCliCfg{DockerApiVersion: "1.40"}
	cli, _ := NewClient(dockerCliCfg)
	if err := cli.RemoveContainer("shuffle-orborus"); err != nil {
		log.Printf("remove Container failed!")
	}
}

/* need expose 2375 port on the remote server*/
func TestConnectOtherHost(t *testing.T) {
	dockerCliCfg := &DockerCliCfg{DockerApiVersion: "1.40", DockerSeverIp: "10.10.15.86", DockerSeverPort: "2375"} // non tls connection
	cli, _ := NewClient(dockerCliCfg)
	dockercli := cli.DockerCli
	images, err := dockercli.ImageList(context.Background(), types.ImageListOptions{})
	_ = images
	if err != nil {
		log.Error(fmt.Sprintf("Unable to list docker images: %s", err))
	}
}

func TestIsContainerRunning(t *testing.T) {
	dockerCliCfg := &DockerCliCfg{DockerApiVersion: "1.40"}
	cli, _ := NewClient(dockerCliCfg)
	if !cli.IsContainerRunning("shuffle-agent") {
		log.Printf("not running!")
	}
}
func deleTarFile(fileName string) error {
	if err := os.Remove(fileName); err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

func TestBuildImage(t *testing.T) {

	var files, dirs []string

	transerveDir := "..\\orborus" //"D:\\code\\go_test"
	if err := utils.GetFilesAndDirs(transerveDir, &files, &dirs); err != nil {
		fmt.Printf("遍历文件夹失败，错误原因：%s", err)
	}
	//_ = dirs
	//files = append(files, "D:\\code\\netsim\\dockerCliWrapper\\orborus\\Dockerfile")
	// need to filter some unuseful files
	dst := "a.tar"
	if err := utils.Tar(files, dst); err != nil {
		log.Fatal(err)
	}
	defer deleTarFile(dst)

	dockerCliCfg := &DockerCliCfg{DockerApiVersion: "1.40"}
	cli, err := NewClient(dockerCliCfg)
	defer cli.Close()
	if err != nil {
		log.Error("Unable to create docker client", err)
	}

	tag := fmt.Sprintf("%s-%s", "app_sdk", "1.0.0")
	buildCfg := BuildCfg{ImageName: tag, TarFile: dst}
	if err := cli.BuildImage(&buildCfg); err != nil {
		log.Debug("Build error:", err.Error())
		return
	}
}
