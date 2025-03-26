## wos-redeem-discord-bot

Whiteout Survival bot auto redeem gift code in Discord

本项目从 [pb8DvwQkfRR/wos-redeem](https://github.com/pb8DvwQkfRR/wos-redeem/) 改进而来，因为变动过大，所以没有在 fork 开发.   
感谢 pb8DvwQkfRR 提供的开源代码！

### 功能

原有功能
- 可以读取文本中的 Code 兑换，格式`Code: xxxxx`,可以用于接收官方 Gift Code 消息自动进行兑换
- 手动输入 /code xxxx 命令进行兑换

本项目在原有项目上的改进
- 增加数据库存储用户信息、熔炉等级
- 在兑换时增加重试机制
- 错误兑换码/过期兑换码取消兑换，减少对 API 的请求
- 同时进行只能进行一个兑换流程，防止冲突，可以使用命令取消兑换
- 如果遇到官方 API 服务器压力多大的情况，可能会出现兑换时重试也失败的情况，建议取消兑换，稍等一会儿再兑换
- 增加应对发送兑换结果的文本超出 discord 限制的处理



### 准备工作

建议使用 ubuntu 系统的 Linux，新手友好，下文教程按 ubuntu 进行

1. supervisor 建议安装用于进程管理，也可以用 systemd
```
apt install supervisor -y
```
2. mongodb 数据库
[cloud.mongodb.com](https://cloud.mongodb.com/) 注册账号，不需要信用卡  
在你的服务器位置最近的地方开一个 free 的实例，可参考教程 https://blog.csdn.net/weixin_44519083/article/details/119610881  
密码设置复杂点  
开好后，在 `Security` -  `Network Access` 中把你自己的 IP 和 服务器的 IP 添加进去，
想省事一点就添加 `0.0.0.0` ， 这就是放行 所有 IP 可以连接  
点击 `Connect`，左边选择 `Compass`，得到数据库连接地址  
其中 `<db_password>` 需要替换成你设置的密码





### 初始化数据库

1. 个人 PC 上下载 mongodb compass
下载链接 https://www.mongodb.com/try/download/compass
选择你的系统下载安装

初始化时不要放超过100个id，具体看用户的昵称长度，不要让长度超过2000个字符，否则会报错，处理起来比较麻烦，这里不想改了，有点头大

2. 将想要兑换的id 放到文件中，比如 ids.csv
表头 fid
内容用户id，一行一个

3. 打开 mongodb compass，点击"+"，在 URI 框粘贴上面得到的数据库连接地址连接
4. 点击 "create database"  
Database Name 输入 wos
Collection Name 输入 user
5. 导入用户ID
选择 `user` 表，点击 `ADD DATA` - `Import JSON or CSV file`，将 `ids.csv` 放入，导入即可

### 部署 bot

1. 在服务器上下载构建好的二进制文件(吐槽下，github action 是真难用，改了七八次才改好, action 经常改版本，用户谁会盯着你更新 action 版本，每次都是报错了才知道，运行一次报错 action 版本老了找不到，再运行一次，说权限又不对了，构建产物传不上去了.....)

下载地址 https://github.com/meoww-bot/wos-redeem-discord-bot/releases/tag/v1

根据你的服务器架构选择对应的文件，如果你的服务器是 Linux，CPU架构是 x86，选择 amd64
这里默认你使用 root 用户
```
cd /root
wget https://github.com/meoww-bot/wos-redeem-discord-bot/releases/download/v1/wos-redeem-discord-bot_1_linux_amd64.tar.gz
mkdir wos-redeem-discord-bot
tar zxvf wos-redeem-discord-bot_1_linux_amd64.tar.gz -C wos-redeem-discord-bot
```
程序已经解压到 /root/wos-redeem-discord-bot/wos-redeem-discord-bot

2. 配置 supervisor
```
wget https://raw.githubusercontent.com/meoww-bot/wos-redeem-discord-bot/refs/heads/master/scripts/wos-redeem-discord-bot.conf
```
在 `wos-redeem-discord-bot.conf` 中，最后一行 `BOT_TOKEN="",MongoURI=""`，引号中对应位置放入 discord bot 的 token 和 mongodb 连接地址

discord bot 如何申请可以谷歌，有很多教程

修改好了后
```
cp wos-redeem-discord-bot.conf /etc/supervisor/conf.d/
supervisorctl reload
```
查看运行状态
```
supervisorctl status
```

查看日志
```
tail -100f /var/log/wos-redeem-discord-bot.log
```

### 在 discord 上操作 bot

bot 正常启动后会自动注册命令，你可以向 bot 发送 `/list` 命令，bot会自动检测数据库中的 id，更新用户信息，可以关注下日志，看看用户 id 是否都是有效的，有没有写错的 id


