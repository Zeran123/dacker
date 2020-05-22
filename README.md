# Dacker

这是一个用于构建有依赖关系的Docker镜像，例如有镜像A，B，C的构建脚本放在同一个Repo中，镜像C依赖镜像B，镜像B又依赖镜像A，那么当镜像A有变动时，会自动构建镜像B和镜像C，如果镜像B有变动，则只会重新构建镜像C。

## 想法

通过JSON配置文件定义镜像之间的依赖关系，例如：

```json
[
		{
			"name": "ubuntu",
			"dockerfile": "./example/ubuntu/dockerfile",
			"image": "ubuntu",
			"tag": "18.04-zh-${BuildNumber}"
		},
		{
			"name": "sidecar",
			"dockerfile": "./example/sidecar/dockerfile",
			"image": "sidecar",
			"tag": "u1804-${BuildNumber}",
			"deps": ["ubuntu"]
		},
		{
			"name": "openjdk-11-jre",
			"dockerfile": "./example/java/openjdk-11-jre/dockerfile",
			"image": "openjdk-11-jre",
			"tag": "${BuildNumber}",
			"deps": ["ubuntu"]
		},
		{
			"name": "openjdk-11-jre-with-sidecar",
			"dockerfile": "./example/java/openjdk-11-jre-with-sidecar/dockerfile",
			"image": "sidecar-java",
			"tag": "openjdk-11-jre-${BuildNumber}",
			"deps": ["ubuntu", "sidecar"]
		}
]
```

当Repo有变更时，对以上每个镜像的构建目录进行如下步骤：

1. 计算构建目录中的各个文件的哈希值，并与之前保存的哈希值进行比较，判断文件是否有变化，如果没有变化，则跳过。
2. 如果文件有变化，则重新构建当前镜像
3. 构建成功后，保存当前镜像构建脚本的新的哈希值，并记录当前镜像的TAG（包含BuildNumber）
4. 检查是否有镜像依赖当前镜像，如果有，则开始构建被依赖的镜像（需替换被依赖镜像的Dockerfile中的镜像TAG）。

用于存储构建过程及结果的JSON数据结构，例如：
```json
[
	{
		"name": "ubuntu",
		"hash": "xxxxx",
		"buildNumber": "20200519.1",
		"image": "ubuntu",
		"tag": "18.04-zh-20200519.1",
		"updatedAt": "2020-05-19 23:11:08"
	},
	{
		"name": "sidecar",
		"hash": "xxxxx",
		"buildNumber": "20200519.1",
		"image": "sidecar",
		"tag": "u1804-20200520.1",
		"updatedAt": "2020-05-19 23:11:08"
	}
]
```

## 步骤

```
1. 读取配置JSON数据
2. 根据镜像依赖关系，输出镜像构建顺序
3. for 镜像构建目录
4.   检查是否已构建
5.   计算构建文件的哈希值
6.   对比哈希值
7.   if 文件有变化
8.      docker build
9.      更新构建版本和tag
10.      获取被依赖的镜像
11.      for 被依赖的镜像
12.         替换目标dockerfile中的image和tag
13.         docker build
14.         更新构建版本和tag
```
