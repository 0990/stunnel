## 多通道配置
当远程主机有多个端口在服务时，希望每个端口都能“被加密”，可以使用多通道配置<br>

假设远程主机有两个tcp服务端口9998,9999,希望都被“加密”<br>
stsever.json可配置为：
```
{
   "tunnels": [
      {
         "authkey": "abcdefg",
         "tcp": {
            "listen": "0.0.0.0:2000",
            "remote": "127.0.0.1:9998"
         }
      }，
      {
         "authkey": "hijklmn",
         "tcp": {
            "listen": "0.0.0.0:2001",
            "remote": "127.0.0.1:9999"
         }
      }
   ],
   "log_level": "info"
}
```
stclient配置：
```
{
   "tunnels": [
      {
         "authkey": "abcdefg",
         "connnum": 10,
         "tcp": {
            "listen": "0.0.0.0:1000",
            "remote": "44.55.66.77:2000"
         }
      }，
      {
         "authkey": "hijklmn",
         "connnum": 10,
         "tcp": {
            "listen": "0.0.0.0:1001",
            "remote": "44.55.66.77:2001"
         }
      }
   ],
   "log_level": "info"
}
```

这样配置启动后，localhost:1000即“伪装”成44.55.66.77:9998服务，localhost:1001即“伪装”成44.55.66.77:9999服务<br>

注意以上配置，不同通道的密钥可以配置不同,但每个通道的客户端服务器密钥要相同<br>

