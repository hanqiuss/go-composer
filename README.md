# go-composer

一个golang写的php包管理器

目前只有安装和生成classmap（默认）功能, 暂不支持命令行`require`、`remove`包

暂时没有自定义`repositories`功能


Usage example
-------
```
go-composer
```
```
go-composer -pro
```


命令行参数
---

```
go-composer -d=../test -lockonly -php=7.2.25 -pro
```

| flag  |                description                   |
|-------|----------------------------------------------|
| `-d=<dir>`         | 指定工作目录，目录下需要有`composer.json`文件   |
| `-lockonly`       | 仅生成 `composer.lock` 文件           |
| `-php=<version>`   | 指定`php`版本            |
| `-pro`             | 不安装require-dev包          |
| `-cp`             | create project 创建项目 需要指定目录 -d 参数         |


待处理问题
---

版本对比工具是用别的语言的改的，并没有完全匹配php的版本规则

目前版本依赖算法提供的是 ```最新版本``` 方案，而php composer 返回的是 ```最少安装``` 方案
