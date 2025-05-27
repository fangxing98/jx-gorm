## kingbase 问题
```shell
1.不支持类型 longtext tinyint
 longtext 替换为 text
  tinyint 替换为 int
2.int不支持指定长度
```

### 获取全部tag
```shell
git tag
```

### 设置tag
```shell
# 简易
git tag v1.0.0

# 附注 Tag（推荐，可记录作者、时间、描述）
git tag -a v1.0.0 -m "Release version 1.0.0"
```

### 推送远程
```shell
# 推送指定tag
git push origin v1.0.0

# 推送本地全部tag
git push origin --tags
```

### 删除tag
```shell
# 删除本地 Tag
git tag -d v1.0.0

# 删除远程 Tag
git push origin --delete v1.0.0
```