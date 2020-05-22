package main

import (
	"io/ioutil"
	"fmt"
	"log"
	"encoding/json"
	"os"
	"os/exec"
	"strings"
	"regexp"
	"path/filepath"
	"bufio"
	"crypto/md5"
)

func main(){
	// 第一个参数为BuildNumber
	buildNumber := os. Args[1]
	log.Printf("本次构建版本号：%s", buildNumber)
	// 从dacker.config中读取镜像配置及依赖关系
	log.Printf("加载配置文件")
	
	images := loadConfig()
	names := sortByDeps(images)

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

		dir := filepath.Dir(image.Dockerfile)
		hashs, modified := hasModified(dir, build.Hash)
		if !modified {
			log.Printf("构建脚本没有变更，跳过构建")
			continue;
		}

		dockerfile, err := ioutil.ReadFile(image.Dockerfile)
		if err != nil {
			log.Fatal("加载Dockerfile文件失败", err)
			return
		}
		
		newfileContent := replacePlaceholders(string(dockerfile))
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
		cmd := exec.Command("sudo", "docker", "build", "-f", newfilePath, "-t", imageName + ":" + tag, dir)
		stdout, _ := cmd.StdoutPipe()
		stderr, _ := cmd.StderrPipe()
		oReader := bufio.NewReader(stdout)
		eReader := bufio.NewReader(stderr)			
		err = cmd.Start()

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
			log.Fatal("构建脚本执行失败", err)
			return
		}

		// 删除临时生成的dockerfile文件
		os.RemoveAll(newfilePath)

		build.Name = image.Name
		build.BuildNumber = buildNumber
		build.Image = imageName 
		build.Tag = tag
		build.Hash = hashs
		_, err = build.SaveBuild()
		if err != nil {
			log.Fatal("保存构建结果失败", err)
		} else {
			log.Printf("已保存 %s 的构建结果", build.Name)
		}
	}
	
}


// 根据镜像的依赖关系计算构建优先级
func sortByDeps(images map[string]Image)([]string){
	// 根据依赖计算出构建的优先级
	var names []string
	priority := make(map[string]int)
	for k, v := range images {
		names = append(names, k)
		deps := v.Deps
		for k := 0; k < len(deps); k++ {
			dep := deps[k]
			_, exists := priority[dep]
			if(!exists){
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
		for j := 0; j < l - 1 - i; j++ {
			if priority[names[j]] < priority[names[j + 1]] {
				names[j], names[j + 1] = names[j + 1], names[j]
			}
		}
	}

	return names;
}

//
// 替换构建脚本中的占位符
// 占位符格式为：${镜像名:配置文件中的字段}
// 例如：${ubuntu:Image}
//
func replacePlaceholders(dockerfile string) string {
	r, err := regexp.Compile(`\$\{(\w|-)+\:(\w|-)+\}`)
	if err != nil {
		fmt.Printf("regex compile error")
		return ""
	}
	// 查找Dockerfile中的占位符，并替换为实际的值
	h := make(map[string]string)
	placeholders := r.FindAllString(string(dockerfile), -1)
	for _, placeholder := range placeholders {
		arr := strings.Split(
			strings.Replace(
				strings.Replace(placeholder, "${", "", -1), "}", "", -1), ":")

		build, err := GetBuild(arr[0])
		if err != nil {
			log.Fatal("load build %s error", arr[0])
		}
		var value string
		switch arr[1] {
		case "Image":
			value = build.Image
		case "Tag":
			value = build.Tag
		}
		h[placeholder] = value
	}
	content := dockerfile
	for k, v := range(h) {
		content = strings.Replace(content, k, v, -1)
	}
	return content
}


//
// 检查镜像构建脚本及相关文件的内容是否有变动
//
func hasModified(dir string, existsHashs map[string]string) (map[string]string, bool) {
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
	content, err := ioutil.ReadFile("./dacker.config")
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