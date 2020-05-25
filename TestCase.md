# 测试用例

## 环境

1. 初始化环境：
	1. 删除data.db文件
	2. 执行 go run . 20200525.1
2. 镜像版本
	1. example.com/ubuntu:18.04-20200525.1
	2. example.com/sidecar:u1804-20200525.1
	3. example.com/openjdk-11-jre:20200525.1
	4. example.com/sidecar-java:openjdk-11-jre-20200525.1

## 场景一：OS构建脚本变更

+ 步骤
1. 修改 ubuntu:18.04-zh/Dockerfile
2. 执行 go run . 20200525.2

+ 预期结果
1. example.com/ubuntu:18.04-20200525.2
2. example.com/sidecar:u1804-20200525.2
3. example.com/openjdk-11-jre:20200525.2
4. example.com/sidecar-java:openjdk-11-jre-20200525.2


## 场景二：Sidecar脚本变更

+ 步骤
1. 修改 sidecar/Dockerfile
2. 执行 go run . 20200525.3

+ 预期结果
1. example.com/ubuntu:18.04-20200525.2
2. example.com/sidecar:u1804-20200525.3
3. example.com/openjdk-11-jre:20200525.2
4. example.com/sidecar-java:openjdk-11-jre-20200525.3

## 场景三：Java脚本变更

+ 步骤
1. 修改 openjdk-11-jre/Dockerfile
2. 执行 go run .20200525.4

+ 预期结果
1. example.com/ubuntu:18.04-20200525.2
2. example.com/sidecar:u1804-20200525.3
3. example.com/openjdk-11-jre:20200525.4
4. example.com/sidecar-java:openjdk-11-jre-20200525.4

## 场景四：Java-With-Sidecar脚本变更

+ 步骤
1. 修改 openjdk-11-jre-with-sidecar/Dockerfile
2. 执行 go run .20200525.5

+ 预期结果
1. example.com/ubuntu:18.04-20200525.2
2. example.com/sidecar:u1804-20200525.3
3. example.com/openjdk-11-jre:20200525.4
4. example.com/sidecar-java:openjdk-11-jre-20200525.5