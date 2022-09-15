https://0.0.0.0:5556/auth/tenxcrd/login?back=/auth?client_id=sample-app&redirect_uri=http%3A%2F%2F0.0.0.0%3A5555%2Fcallback&response_type=code&scope=openid+profile+email+offline_access&state=I+wish+to+wash+my+irish+wristwatch&state=e3ynbyhy3gvgngqprptcfluvg



## 镜像构建
在 macos 上直接通过 CGOENABLE=0  无法本地构建，依赖于 sqlit 的包，只能在 容器中构建。

当构建时出现

```
WARNING: Ignoring https://dl-cdn.alpinelinux.org/alpine/v3.14/community: temporary error (try again later)
```
是 docker 网络的问题，
1. 可以通过 在 Dockerfile 中添加代理
2. docker build 时增加 --network=host，使用 host 网络构建


## 自建 reqeust 方式

``` yaml
kind: AuthRequest
apiVersion: dex.coreos.com/v1
metadata:
  name: ve22u653rzr2tjtq3najgt76s
  namespace: default
  creationTimestamp:
clientID: sample-app
responseTypes:
- code
scopes:
- openid
- profile
- email
- offline_access
redirectURI: http://0.0.0.0:5555/callback
state: I wish to wash my irish wristwatch
loggedIn: true
claims:
  userID: ''
  username: ''
  preferredUsername: ''
  email: ''
  emailVerified: false
connectorID: k8scrd
expiry: '2022-05-24T07:46:39.304739Z'
code_challenge_method: plain
```

``` yaml
kind: AuthRequest
apiVersion: dex.coreos.com/v1
metadata:
  name: p75t7qrud2g6y4ilsqptqug57
  namespace: default
  creationTimestamp:
clientID: sample-app
responseTypes:
- code
scopes:
- openid
- profile
- email
- offline_access
redirectURI: http://0.0.0.0:5555/callback
state: I wish to wash my irish wristwatch
loggedIn: false
claims:
  userID: ''
  username: ''
  preferredUsername: ''
  email: ''
  emailVerified: false
connectorID: k8scrd
expiry: '2022-05-24T08:22:56.833702Z'
code_challenge_method: plain

```
