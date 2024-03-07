# <center>山与路远程控制<center>

# 🎥项目演示地址

还在制作....

# ♻️项目基本介绍

山与路远程控制是基于`electron(vue3)`和`golang`实现的远程控制软件(项目界面主要模仿向日葵远程软件,如有侵权请告知),代码可能有点臃肿毕竟只花了一周左右写的无聊项目,如果对其感兴趣的大佬可以fork自行修改(大佬勿喷)
## 技术栈
1. vue3
2. golang
3. webrtc
4. cgo
5. turn
## 项目思路
* 键鼠主要通过cgo调用`windowapi`和`goreboot`来实现击键的监听和控制端模拟
* 控制端和被控制根据服务器转发后在通过`webrtc`和golang的`pion`实现p2p通信
* 服务端主要功能是为了`转发数据`和提供`turn服务器`
* 视频流直接使用最简单的`webrtc`实现实时画面传输
* 当新设备访问服务端则会自动分配识别码和验证码,通过识别码和验证码来建立连接


# 🧧 作者自己的配置环境

## nodejs

`16.20.1`

## npm

![在这里插入图片描述](https://img-blog.csdnimg.cn/732b1f4872104f28955cfdab601bf0c8.png)
## golang

![在这里插入图片描述](https://img-blog.csdnimg.cn/direct/3437d3f67c7640cd8dc74a7ed773e51f.png)
# 📍 服务端相关配置config.yaml
```yaml
serveraddr: ":9998"  #后端地址
turn:
  public_ip: "127.0.0.1" #公网地址
  port: 3478  # turn端口
  thread_num: 5  # 如果服务端是linux则填写   进程数
db_path: "./remote.db" # 数据库路径

```
# 🔖项目运行

## 后端运行

```shell
git clone https://github.com/OblivionTime/remote_electron.git
cd /remote_electron/server
go mod tidy
go run main.go
```

## 前端调试运行

```shell
git clone https://github.com/OblivionTime/remote_electron.git
#启动客户端
cd /remote_electron/client
go mod tidy
go run main.go

# 启动前端页面
cd /remote_electron/ui
yarn
#调试
yarn serve
#打包
yarn build
```





# 👻注意事项

1. 打开软件第一时间修改服务器地址

![在这里插入图片描述](https://img-blog.csdnimg.cn/direct/8333eed7c80f4526bdcbce7d8bc722f5.png)

**技术人员根据自己的需求去修改**

# 🎉已完成功能

* 获取识别码和验证码
* webrtc 建连
* 连接后完全控制对方的键鼠
* 断开连接
* 悬浮球
* 保存连接过的设备
* 实现自己的turn服务器
* 文件传输功能
# 🖼️ 项目截图

![在这里插入图片描述](https://img-blog.csdnimg.cn/direct/bbd01349c078467eabae7c8932b48b6d.png)
![在这里插入图片描述](https://img-blog.csdnimg.cn/direct/9fb5fa81c3dd429d978208be0dbed147.png)
![在这里插入图片描述](https://img-blog.csdnimg.cn/direct/c4ab491431304ead91913ee6b3f627a8.png)
![在这里插入图片描述](https://img-blog.csdnimg.cn/direct/7f574cc3887a45e58a1543e73f847f45.png)
![在这里插入图片描述](https://img-blog.csdnimg.cn/direct/9b0e815f5b5e44a8aab2a7ee755bc874.png)
![在这里插入图片描述](https://img-blog.csdnimg.cn/direct/42faa84b55a8410d9b7e3bb19287afd4.png)
![在这里插入图片描述](https://img-blog.csdnimg.cn/direct/82128b1fd6894065a1170506ccd2497a.png)

![在这里插入图片描述](https://img-blog.csdnimg.cn/direct/693e8a2cb66649daa298ac7811498c9a.png)
![在这里插入图片描述](https://img-blog.csdnimg.cn/direct/f8aaa5761f7f4d048a51390271f3a034.png)

## 注意electron打包会出现的问题

## 打包前必须做的事

- 进入到下面目录C:\Users\自己的用户名\AppData\Local\electron-builder\Cache
  ![在这里插入图片描述](https://img-blog.csdnimg.cn/07e7a371077042039fe75a7aae4ada23.png)



- 创建目录**winCodeSign**和**nsis**
  ![在这里插入图片描述](https://img-blog.csdnimg.cn/99b0ca3b41fb424498775ad81274c950.png)



- 将**electron必须安装包**目录下的**winCodeSign-2.6.0.7z**解压到**C:\Users\自己的用户名\AppData\Local\electron-builder\Cache\winCodeSign**目录下

![在这里插入图片描述](https://img-blog.csdnimg.cn/fb7781236a1b4fa8a8b42f2c19c80346.png)


- 进入到**C:\Users\自己的用户名\AppData\Local\electron-builder\Cache\nsis**目录下,将**electron必须安装包**目录下分别解压成如下图所示的样子
  ![在这里插入图片描述](https://img-blog.csdnimg.cn/6d91296313c9490a9de1b58c0db6373e.png)
  详细教程:[https://www.cnblogs.com/liliyou/p/13423709.html](https://www.cnblogs.com/liliyou/p/13423709.html)



# 结语


![在这里插入图片描述](https://img-blog.csdnimg.cn/e8be97b67c1b43a68add5c6c5944fbc9.jpeg)

