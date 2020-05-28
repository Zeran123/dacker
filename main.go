package main

import (
	"bufio"
	"crypto/md5"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

/*
是否需要推送镜像
*/
var push bool

/*
本次构建的版本号
*/
var buildNumber string

func main() {
	buildNumber = ""
	push = false
	flag.BoolVar(&push, "push", false, "开启镜像推送")
	flag.StringVar(&buildNumber, "v", "", "设置构建版本号")
	flag.Parse()

	len_args := len(os.Args)

	if len_args == 1 {
		fmt.Printf("Usage: dacker [arguments] command\n")
		fmt.Printf("The arguments are:\n")
		fmt.Printf("\t-push=bool\t是否推送镜像，默认false\n")
		fmt.Printf("\t-v=string\t构建版本号\n")
		fmt.Printf("\n")
		fmt.Printf("The commands are:\n")
		fmt.Printf("\tbuild\t构建镜像,需配合 -v 参数使用\n")
		fmt.Printf("\trelease\t发布镜像\n")
		fmt.Printf("\tlog\t查看镜像构建记录\n")
		return
	}

	op := strings.ToLower(os.Args[len_args-1])
	switch op {
	case "build":
		{
			if buildNumber == "" {
				fmt.Println("missing '-v buildNumber'")
				return
			}
			doBuild()
		}
	case "release":
		{
			doRelease()
		}
	case "log":
		{
			doLog()
		}
	default:
		{
			fmt.Println("Unknown Command")
		}
	}

}

//
// 构建
func doBuild() {
	log.Printf("本次构建版本号：%s", buildNumber)
	// 从dacker.config中读取镜像配置及依赖关系
	log.Printf("加载配置文件")

	images := loadConfig()
	names := sortByDeps(images)
	for _, n := range names {
		log.Printf("%s ", n)
	}
	log.Printf("共加载镜像 %d 个", len(names))

	for _, name := range names {
		log.Printf("==================================")

		image := images[name]
		log.Printf("准备构建镜像 %s", name)
		build, err := GetBuild(image.Name)
		if err != nil {
			build = Buildlog{}
		}
		_, err = os.Stat(image.Dockerfile)
		if err != nil {
			log.Printf("Dockerfile文件不存在，跳过")
			continue
		}

		dockerfile, err := ioutil.ReadFile(image.Dockerfile)
		if err != nil {
			log.Fatal("加载Dockerfile文件失败", err)
			return
		}

		/*
			检查当前镜像的依赖的Tag是否有变动
		*/
		var depNames []string
		depModified := false
		p, v := getDependency(string(dockerfile))
		/*
			v 为当前镜像dockerfile中依赖的镜像名称和tag
		*/
		for n, t := range build.Deps {
			if v[n] != t {
				depNames = append(depNames, n+":"+t)
				depModified = true
			}
		}

		dir := filepath.Dir(image.Dockerfile)
		hashs, modified := isModified(dir, build.Hash)
		if depModified {
			log.Printf("依赖镜像 ")
			for _, n := range depNames {
				log.Printf("%s ", n)
			}
			log.Printf("发生变更")
		} else {
			if !modified {
				log.Printf("构建脚本没有变更，跳过构建")
				continue
			}
		}

		newfileContent := replacePlaceholders(string(dockerfile), p)
		newfilePath := image.Dockerfile + "_active"

		// 删除临时生成的dockerfile文件
		os.RemoveAll(newfilePath)

		err = ioutil.WriteFile(newfilePath, []byte(newfileContent), 0666)
		if err != nil {
			log.Fatal("保存Dockerfile文件失败", err)
			return
		}

		// exists
		imageName := strings.Replace(image.Image, "${BuildNumber}", buildNumber, -1)
		tag := strings.Replace(image.Tag, "${BuildNumber}", buildNumber, -1)

		// 执行 docker build 构建命令
		log.Printf("开始构建镜像 => %s:%s", imageName, tag)
		invokeCmd("sudo", "docker", "build", "-f", newfilePath, "-t", imageName+":"+tag, dir)

		// 删除临时生成的dockerfile文件
		os.RemoveAll(newfilePath)

		build.Name = image.Name
		build.BuildNumber = buildNumber
		build.Image = imageName
		build.Tag = tag
		build.Hash = hashs
		build.Deps = make(map[string]string)
		for n, t := range v {
			build.Deps[n] = t
		}
		_, err = build.SaveBuild()
		if err != nil {
			log.Fatal("保存构建结果失败", err)
		} else {
			log.Printf("已保存 %s 的构建结果", build.Name)
			if push {
				invokeCmd("sudo", "docker", "push", imageName+":"+tag)
			}
		}
	}
}

//
// 发布
func doRelease() {
	builds, err := ListBuild()
	if err != nil {
		log.Fatal("获取构建记录失败, %v", err)
		return
	}

	images := loadConfig()

	for _, build := range builds {
		if build.Tag == build.ReleaseTag {
			continue
		}

		log.Printf("镜像 %s 的 ReleaseTag 为 %s，当前最新的 Tag 为 %s", build.Name, build.ReleaseRef, build.Tag)

		// ReTag
		image := build.Image
		tag := build.Tag
		releaseTag := images[build.Name].Release
		log.Println(releaseTag)
		lastBuildImage := image + ":" + tag
		releaseImage := image + ":" + releaseTag

		invokeCmd("sudo", "docker", "tag", lastBuildImage, releaseImage)
		if push {
			invokeCmd("sudo", "docker", "push", releaseImage)
		}

		build.ReleaseTag = releaseTag
		build.ReleaseRef = tag
		_, err = build.SaveBuild()
		if err != nil {
			log.Fatal("保存Release结果失败", err)
		}
	}
}

//
// 查看构建记录
func doLog() {
	builds, err := ListBuild()
	if err != nil {
		log.Fatal("获取构建记录失败, %v", err)
		return
	}

	for _, build := range builds {
		fmt.Printf("Name: %s\n", build.Name)
		fmt.Printf("BuildNumber: %s\n", build.BuildNumber)
		fmt.Printf("Image: %s\n", build.Image)
		fmt.Printf("BuildTag: %s\n", build.Tag)
		fmt.Printf("ReleaseTag: %s\n", build.ReleaseTag)
		fmt.Printf("ReleaseRef: %s\n", build.ReleaseRef)
		fmt.Printf("=======================================\n")
	}
}

//
// 调用本地命令，并同步输出日志
//
func invokeCmd(name string, arg ...string) {
	cmd := exec.Command(name, arg...)
	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()
	oReader := bufio.NewReader(stdout)
	eReader := bufio.NewReader(stderr)
	err := cmd.Start()

	// 同步输出日志
	go func() {
		for {
			line, _ := oReader.ReadString('\n')
			if line != "" {
				log.Printf(line)
			}
		}
	}()

	// 同步输出错误
	go func() {
		for {
			line, _ := eReader.ReadString('\n')
			if line != "" {
				log.Printf(line)
			}
		}
	}()

	err = cmd.Wait()
	if err != nil {
		log.Fatal(err)
		return
	}
}

// 根据镜像的依赖关系计算构建优先级
func sortByDeps(images map[string]Image) []string {
	// 根据依赖计算出构建的优先级
	var names []string
	priority := make(map[string]int)
	for k, v := range images {
		names = append(names, k)
		deps := v.Deps
		for k := 0; k < len(deps); k++ {
			dep := deps[k]
			_, exists := priority[dep]
			if !exists {
				priority[dep] = 1
			} else {
				priority[dep] = priority[dep] + 1
			}
		}
	}

	// 按被依赖的数量倒序排序
	// 构建镜像的顺序
	l := len(names)
	for i := 0; i < l; i++ {
		for j := 0; j < l-1-i; j++ {
			if priority[names[j]] < priority[names[j+1]] {
				names[j], names[j+1] = names[j+1], names[j]
			}
		}
	}

	return names
}

//
// 从dockerfile中解析依赖镜像的占位符
// 返回两个结果：
// 1. map[占位符]=实际的值
// 2. map[依赖的镜像名]=依赖镜像的Tag
//
func getDependency(dockerfile string) (map[string]string, map[string]string) {
	// 占位符
	h := make(map[string]string)
	// 镜像版本
	v := make(map[string]string)
	r, _ := regexp.Compile(`\$\{(\w|-)+\:(\w|-)+\}`)
	placeholders := r.FindAllString(string(dockerfile), -1)
	/*
		解析dockerfile中的占位符${依赖镜像的Name:依赖镜像的Tag}
		然后从当前构建的结果中获取依赖镜像的最新的Tag
	*/
	for _, placeholder := range placeholders {
		arr := strings.Split(
			strings.Replace(
				strings.Replace(placeholder, "${", "", -1), "}", "", -1), ":")

		name := arr[0]
		build, err := GetBuild(name)
		if err != nil {
			log.Fatal("load build %s error", arr[0])
		}
		var field string
		switch arr[1] {
		case "Image":
			field = build.Image
		case "Tag":
			field = build.Tag
			v[name] = field
		}
		h[placeholder] = field
	}
	return h, v
}

//
// 替换构建脚本中的占位符
// 占位符格式为：${镜像名:配置文件中的字段}
// 例如：${ubuntu:Image}
//
func replacePlaceholders(dockerfile string, p map[string]string) string {
	// 查找Dockerfile中的占位符，并替换为实际的值
	content := dockerfile
	for k, v := range p {
		content = strings.Replace(content, k, v, -1)
	}
	return content
}

//
// 检查镜像构建脚本及相关文件的内容是否有变动
//
func isModified(dir string, existsHashs map[string]string) (map[string]string, bool) {
	// 计算构建脚本所在目录中所有文件的Hash值
	log.Printf("准备计算构建脚本的哈希值 => %s", dir)
	hashs := make(map[string]string)
	files, _ := ioutil.ReadDir(dir)
	for _, file := range files {
		fileName := file.Name()
		log.Printf("开始计算文件的哈希值 => %s", fileName)
		content, err := ioutil.ReadFile(filepath.Join(dir, fileName))
		if err != nil {
			log.Fatal(err)
		}
		hash := md5.Sum(content)
		hashStr := fmt.Sprintf("%x", hash)
		hashs[fileName] = string(hashStr)
	}

	modified := false
	for k, v := range hashs {
		oldValue, exists := existsHashs[k]
		if !exists || v != oldValue {
			modified = true
		}
	}

	return hashs, modified
}

//
// 加载配置文件
//
func loadConfig() map[string]Image {
	content, err := ioutil.ReadFile("./dacker.conf")
	if err != nil {
		log.Fatal(err)
	}

	var imagesArr []Image
	err = json.Unmarshal(content, &imagesArr)
	if err != nil {
		log.Fatal(err)
	}
	imagesMap := make(map[string]Image)
	for _, image := range imagesArr {
		imagesMap[image.Name] = image
	}

	return imagesMap
}
